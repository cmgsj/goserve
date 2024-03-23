package files

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"
)

type Server struct {
	fs           fs.FS
	exclude      *regexp.Regexp
	handlers     map[string]Handler
	contentTypes []string
}

func NewServer(fs fs.FS, exclude *regexp.Regexp, handlers ...Handler) *Server {
	server := &Server{
		fs:       fs,
		exclude:  exclude,
		handlers: make(map[string]Handler),
	}

	for _, handler := range handlers {
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
			s.handleError(w, nil, newUnsupportedContentTypeError(contentType, s.contentTypes), http.StatusBadRequest)
			return
		}

		file = path.Clean(file)

		info, err := fs.Stat(s.fs, file)
		if err != nil {
			s.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		if !s.IsAllowed(file) {
			s.handleError(w, handler, newFileNotFoundError(file), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = s.copyFile(w, file)
			if err != nil {
				s.handleError(w, handler, err, fsErrorStatusCode(err))
			}
			return
		}

		entries, err := s.readDir(file)
		if err != nil {
			s.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		err = handler.HandleDir(w, file, entries)
		if err != nil {
			s.handleError(w, handler, err, http.StatusInternalServerError)
		}
	})
}

func (s *Server) ContentTypesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, s.contentTypes)
	})
}

func (s *Server) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func (s *Server) ContentTypes() []string {
	return s.contentTypes
}

func (s *Server) IsAllowed(file string) bool {
	if file == RootDir || s.exclude == nil {
		return true
	}

	for _, path := range strings.Split(file, "/") {
		if s.exclude.MatchString(path) {
			return false
		}
	}

	return true
}

func (s *Server) copyFile(w http.ResponseWriter, file string) error {
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

func (s *Server) readDir(dir string) ([]File, error) {
	entries, err := fs.ReadDir(s.fs, dir)
	if err != nil {
		return nil, err
	}

	var files []File

	if dir != RootDir {
		files = append(files, File{
			Path:  path.Dir(dir),
			Name:  ParentDir,
			IsDir: true,
		})
	}

	for _, entry := range entries {
		file := path.Join(dir, entry.Name())

		if !s.IsAllowed(file) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		files = append(files, File{
			Path:  file,
			Name:  info.Name(),
			Size:  FormatSize(info.Size()),
			IsDir: info.IsDir(),
		})
	}

	Sort(files)

	return files, nil
}

func (s *Server) handleError(w http.ResponseWriter, handler Handler, err error, code int) {
	slog.Error("an error ocurred", "error", err)

	w.WriteHeader(code)

	if handler != nil {
		handleErr := handler.HandleError(w, err, code)
		if handleErr == nil {
			return
		}
		slog.Error("failed to handle error", "error", handleErr)
	}

	fmt.Fprintln(w, err.Error())
}
