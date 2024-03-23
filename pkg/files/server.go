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
			s.SendError(w, newUnsupportedContentTypeError(s.contentTypes, contentType), http.StatusBadRequest)
			return
		}

		file = path.Clean(file)

		info, err := fs.Stat(s.fs, file)
		if err != nil {
			switch {
			case errors.Is(err, fs.ErrNotExist):
				handler.SendError(w, err, http.StatusNotFound)
			case errors.Is(err, fs.ErrPermission):
				handler.SendError(w, err, http.StatusUnauthorized)
			default:
				handler.SendError(w, err, http.StatusInternalServerError)
			}
			return
		}

		if !s.IsAllowed(file) {
			handler.SendError(w, newFileNotFoundError(file), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = s.sendFile(w, file)
			if err != nil {
				handler.SendError(w, err, http.StatusInternalServerError)
			}
			return
		}

		entries, err := fs.ReadDir(s.fs, file)
		if err != nil {
			handler.SendError(w, err, http.StatusInternalServerError)
			return
		}

		err = handler.SendDir(w, file, entries)
		if err != nil {
			handler.SendError(w, err, http.StatusInternalServerError)
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
	name := path.Base(file)

	if name == RootDir {
		return true
	}

	if !s.dotfiles {
		return !strings.HasPrefix(name, ".")
	}

	return true
}

func (s *Server) SendError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
	slog.Error("an error ocurred", "error", err)
}

func (s *Server) sendFile(w http.ResponseWriter, file string) error {
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
