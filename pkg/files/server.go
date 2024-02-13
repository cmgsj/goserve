package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/cmgsj/goserve/pkg/templates"
	utilbytes "github.com/cmgsj/goserve/pkg/util/bytes"
)

const (
	rootDir   = "."
	parentDir = ".."

	pathValueFile = "file"
)

type Server struct {
	fs.FS
	includeDotfiles bool
	version         string
}

func NewServer(fs fs.FS, includeDotfiles bool, version string) *Server {
	return &Server{
		FS:              fs,
		includeDotfiles: includeDotfiles,
		version:         version,
	}
}

func (s *Server) ServePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := r.PathValue(pathValueFile)

		file = path.Clean(file)

		info, err := fs.Stat(s, file)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				s.sendErrorPage(w, err, http.StatusNotFound)
			} else {
				s.sendErrorPage(w, err, http.StatusInternalServerError)
			}
			return
		}

		if !s.isAllowed(file) {
			s.sendErrorPage(w, errFileNotFound(file), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = s.sendFile(w, file)
			if err != nil {
				s.sendErrorPage(w, err, http.StatusInternalServerError)
			}
			return
		}

		entries, err := fs.ReadDir(s, file)
		if err != nil {
			s.sendErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		err = s.sendPage(w, entries, file)
		if err != nil {
			s.sendErrorPage(w, err, http.StatusInternalServerError)
		}
	})
}

func (s *Server) ServeText() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := r.PathValue(pathValueFile)

		file = path.Clean(file)

		info, err := fs.Stat(s, file)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				sendError(w, err, http.StatusNotFound)
			} else {
				sendError(w, err, http.StatusInternalServerError)
			}
			return
		}

		if !s.isAllowed(file) {
			sendError(w, errFileNotFound(file), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = s.sendFile(w, file)
			if err != nil {
				sendError(w, err, http.StatusInternalServerError)
			}
			return
		}

		entries, err := fs.ReadDir(s, file)
		if err != nil {
			sendError(w, err, http.StatusInternalServerError)
			return
		}

		err = s.sendText(w, entries, file)
		if err != nil {
			sendError(w, err, http.StatusInternalServerError)
		}
	})
}

func (s *Server) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (s *Server) Version() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, s.version)
	})
}

func (s *Server) sendFile(w http.ResponseWriter, file string) error {
	f, err := s.Open(file)
	if err != nil {
		return err
	}

	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("failed to close file", "file", file, "error", err)
		}
	}()

	_, err = io.Copy(w, f)

	return err
}

func (s *Server) sendPage(w http.ResponseWriter, entries []fs.DirEntry, file string) error {
	var breadcrumbs, files []templates.File

	if file != rootDir {
		var pathPrefix string

		for _, name := range strings.Split(file, "/") {
			pathPrefix = path.Join(pathPrefix, name)

			breadcrumbs = append(breadcrumbs, templates.File{
				Path: pathPrefix,
				Name: name,
			})
		}

		files = append(files, templates.File{
			Path:  path.Dir(file),
			Name:  parentDir,
			IsDir: true,
		})
	}

	for _, entry := range entries {
		if !s.isAllowed(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		files = append(files, templates.File{
			Path:  path.Join(file, info.Name()),
			Name:  info.Name(),
			Size:  utilbytes.FormatSize(info.Size()),
			IsDir: info.IsDir(),
		})
	}

	templates.SortFiles(files)

	return templates.ExecuteIndex(w, templates.Index{
		Breadcrumbs: breadcrumbs,
		Files:       files,
		Version:     s.version,
	})
}

func (s *Server) sendText(w io.Writer, entries []fs.DirEntry, file string) error {
	var files []templates.File

	if file != rootDir {
		files = append(files, templates.File{
			Name:  parentDir,
			IsDir: true,
		})
	}

	for _, entry := range entries {
		if !s.isAllowed(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		files = append(files, templates.File{
			Name:  info.Name(),
			Size:  utilbytes.FormatSize(info.Size()),
			IsDir: info.IsDir(),
		})
	}

	templates.SortFiles(files)

	var buf bytes.Buffer

	for _, file := range files {
		buf.WriteString(file.Name)
		if file.IsDir {
			buf.WriteByte('/')
		} else {
			buf.WriteByte('\t')
			buf.WriteString(file.Size)
		}
		buf.WriteByte('\n')
	}

	tab := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	defer tab.Flush()

	_, err := io.Copy(tab, &buf)

	return err
}

func (s *Server) isAllowed(file string) bool {
	name := path.Base(file)

	if name == rootDir {
		return true
	}

	if !s.includeDotfiles {
		return !strings.HasPrefix(name, ".")
	}

	return true
}

func (s *Server) sendErrorPage(w http.ResponseWriter, err error, code int) {
	err = templates.ExecuteIndex(w, templates.Index{
		Error: &templates.Error{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
		Version: s.version,
	})
	if err != nil {
		sendError(w, err, http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
	slog.Error("an error ocurred", "error", err)
}

func errFileNotFound(file string) error {
	return &fs.PathError{
		Op:   "open",
		Path: file,
		Err:  syscall.ENOENT,
	}
}
