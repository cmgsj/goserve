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
	"text/tabwriter"

	"github.com/cmgsj/goserve/pkg/templates"
	"github.com/cmgsj/goserve/pkg/util/units"
)

const (
	rootDir   = "."
	parentDir = ".."
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
		filepath := path.Clean(r.PathValue("path"))

		info, err := fs.Stat(s, filepath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				s.sendErrorPage(w, err, http.StatusNotFound)
			} else {
				s.sendErrorPage(w, err, http.StatusInternalServerError)
			}
			return
		}

		if !s.isAllowed(filepath) {
			s.sendErrorPage(w, errOpenNoSuchFileOrDirectory(filepath), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = s.sendFile(w, r, filepath)
			if err != nil {
				s.sendErrorPage(w, err, http.StatusInternalServerError)
			}
			return
		}

		entries, err := fs.ReadDir(s, filepath)
		if err != nil {
			s.sendErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		err = s.sendPage(w, entries, filepath)
		if err != nil {
			s.sendErrorPage(w, err, http.StatusInternalServerError)
		}
	})
}

func (s *Server) ServeText() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filepath := path.Clean(r.PathValue("path"))

		info, err := fs.Stat(s, filepath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				sendError(w, err, http.StatusNotFound)
			} else {
				sendError(w, err, http.StatusInternalServerError)
			}
			return
		}

		if !s.isAllowed(filepath) {
			sendError(w, errOpenNoSuchFileOrDirectory(filepath), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = s.sendFile(w, r, filepath)
			if err != nil {
				sendError(w, err, http.StatusInternalServerError)
			}
			return
		}

		entries, err := fs.ReadDir(s, filepath)
		if err != nil {
			sendError(w, err, http.StatusInternalServerError)
			return
		}

		err = s.sendText(w, entries, filepath)
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

func (s *Server) sendFile(w http.ResponseWriter, r *http.Request, filepath string) error {
	f, err := s.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)

	return err
}

func (s *Server) sendPage(w http.ResponseWriter, entries []fs.DirEntry, filepath string) error {
	var breadcrumbs, files []templates.File

	if filepath != rootDir {
		var pathPrefix string

		for _, name := range strings.Split(filepath, "/") {
			pathPrefix = path.Join(pathPrefix, name)

			breadcrumbs = append(breadcrumbs, templates.File{
				Path: pathPrefix,
				Name: name,
			})
		}

		files = append(files, templates.File{
			Path:  path.Dir(filepath),
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
			Path:  path.Join(filepath, info.Name()),
			Name:  info.Name(),
			Size:  units.FormatSize(info.Size()),
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

func (s *Server) sendText(w io.Writer, entries []fs.DirEntry, filepath string) error {
	var files []templates.File

	if filepath != rootDir {
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
			Size:  units.FormatSize(info.Size()),
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

func (s *Server) isAllowed(filepath string) bool {
	name := path.Base(filepath)
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

func errOpenNoSuchFileOrDirectory(filepath string) error {
	return &fs.PathError{
		Op:   "open",
		Path: filepath,
		Err:  fs.ErrNotExist,
	}
}
