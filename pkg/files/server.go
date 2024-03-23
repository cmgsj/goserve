package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strings"
)

type Server struct {
	fs           fs.FS
	dotfiles     bool
	version      string
	handlers     map[string]Handler
	contentTypes []string
}

func NewServer(fs fs.FS, dotfiles bool, version string, factories ...HandlerFactory) *Server {
	server := &Server{
		fs:       fs,
		dotfiles: dotfiles,
		version:  version,
		handlers: make(map[string]Handler),
	}

	for _, factory := range factories {
		handler := factory(server)
		server.handlers[handler.ContentType()] = handler
		server.contentTypes = append(server.contentTypes, handler.ContentType())
	}

	slices.Sort(server.contentTypes)

	return server
}

func (s *Server) FilesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.PathValue("content_type")
		file := r.PathValue("file")

		handler, ok := s.handlers[contentType]
		if !ok {
			s.HandleError(w, newUnsupportedContentTypeError(s.contentTypes, contentType), http.StatusBadRequest)
			return
		}

		file = path.Clean(file)

		info, err := fs.Stat(s.fs, file)
		if err != nil {
			handler.HandleError(w, err, fsErrorStatusCode(err))
			return
		}

		if !s.IsAllowed(file) {
			handler.HandleError(w, newFileNotFoundError(file), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = s.handleFile(w, file)
			if err != nil {
				handler.HandleError(w, err, fsErrorStatusCode(err))
			}
			return
		}

		entries, err := fs.ReadDir(s.fs, file)
		if err != nil {
			handler.HandleError(w, err, fsErrorStatusCode(err))
			return
		}

		err = handler.HandleDir(w, file, entries)
		if err != nil {
			handler.HandleError(w, err, http.StatusInternalServerError)
		}
	})
}

func (s *Server) ContentTypesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, s.contentTypes)
	})
}

func (s *Server) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (s *Server) VersionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, s.version)
	})
}

func (s *Server) ContentTypes() []string {
	return s.contentTypes
}

func (s *Server) Version() string {
	return s.version
}

func (s *Server) IsAllowed(file string) bool {
	if file == RootDir {
		return true
	}

	for _, path := range strings.Split(file, "/") {
		if !s.dotfiles && strings.HasPrefix(path, ".") {
			return false
		}
	}

	return true
}

func (s *Server) HandleError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
	slog.Error("an error ocurred", "error", err)
}

func (s *Server) handleFile(w http.ResponseWriter, file string) error {
	f, err := s.fs.Open(file)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			slog.Error("failed to close file", "file", file, "error", err)
		}
	}()

	_, err = io.Copy(w, f)

	return err
}

func newUnsupportedContentTypeError(contentTypes []string, contentType string) error {
	return fmt.Errorf("unsupported content type %q, supported: [%s]", contentType, strings.Join(contentTypes, ","))
}

func newFileNotFoundError(file string) error {
	return fmt.Errorf("stat %s: no such file or directory", file)
}

func fsErrorStatusCode(err error) int {
	switch {
	case errors.Is(err, fs.ErrInvalid):
		return http.StatusBadRequest

	case errors.Is(err, fs.ErrPermission):
		return http.StatusUnauthorized

	case errors.Is(err, fs.ErrNotExist):
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}
