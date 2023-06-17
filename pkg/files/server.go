package files

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cmgsj/goserve/pkg/templates"
	"github.com/spf13/afero"
)

type Server struct {
	Fs           afero.Fs
	Stderr       io.Writer
	SkipDotFiles bool
	RawEnabled   bool
}

func NewServer(fsys afero.Fs, stdout, stderr io.Writer, skipDotFiles, rawEnabled, logEnabled bool) http.Handler {
	s := &Server{
		Fs:           fsys,
		Stderr:       stderr,
		SkipDotFiles: skipDotFiles,
		RawEnabled:   rawEnabled,
	}
	if logEnabled {
		return RequestLogger(s, stdout)
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath := path.Clean(r.URL.Path)
	info, err := s.Fs.Stat(filePath)
	if err != nil {
		s.sendErrorPage(w, err)
		return
	}
	if info.IsDir() {
		dir, err := afero.ReadDir(s.Fs, filePath)
		if err != nil {
			s.sendErrorPage(w, err)
			return
		}
		var files []templates.File
		for _, info := range dir {
			name := info.Name()
			if s.SkipDotFiles && strings.HasPrefix(name, ".") {
				continue
			}
			files = append(files, templates.File{
				Path:  (&url.URL{Path: name}).String(),
				Name:  templates.ReplaceHTML(name),
				Size:  formatFileSize(info.Size()),
				IsDir: info.IsDir(),
			})
		}
		sort.Sort(FileSlice(files))
		err = templates.ExecuteIndex(w, templates.Page{
			Ok:       true,
			BackLink: path.Dir(filePath),
			Header:   filePath,
			Files:    files,
		})
		if err != nil {
			s.sendError(w, err)
		}
	} else {
		s.sendFile(w, r, filePath)
	}
}

func (s *Server) sendErrorPage(w http.ResponseWriter, err error) {
	err = templates.ExecuteIndex(w, templates.Page{
		Ok:       false,
		BackLink: "/",
		Header:   err.Error(),
	})
	if err != nil {
		s.sendError(w, err)
	}
}

func (s *Server) sendFile(w http.ResponseWriter, r *http.Request, filePath string) {
	f, err := s.Fs.Open(filePath)
	if err != nil {
		s.sendError(w, err)
		return
	}
	defer f.Close()
	if !s.RawEnabled {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		s.sendError(w, err)
		return
	}
	r.Header.Set(BytesCopiedHeader, formatFileSize(n))
}

func (s *Server) sendError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	fmt.Fprintln(s.Stderr, err)
}
