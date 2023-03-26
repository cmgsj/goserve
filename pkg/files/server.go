package files

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/cmgsj/goserve/pkg/templates"
	"github.com/spf13/afero"
)

type Server struct {
	fs           afero.Fs
	errC         chan<- error
	skipDotFiles bool
	rawEnabled   bool
	version      string
}

type ServerConfig struct {
	Fs           afero.Fs
	ErrC         chan<- error
	SkipDotFiles bool
	RawEnabled   bool
	Version      string
}

func NewServer(config ServerConfig) http.Handler {
	return &Server{
		fs:           config.Fs,
		errC:         config.ErrC,
		skipDotFiles: config.SkipDotFiles,
		rawEnabled:   config.RawEnabled,
		version:      config.Version,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath := path.Clean(r.URL.Path)
	info, err := s.fs.Stat(filePath)
	if err != nil {
		s.sendErrorPage(w, err)
		return
	}
	if info.IsDir() {
		filesInfo, err := afero.ReadDir(s.fs, filePath)
		if err != nil {
			s.sendErrorPage(w, err)
			return
		}
		var dirs, files []templates.File
		for _, fileInfo := range filesInfo {
			if s.skipDotFiles && strings.HasPrefix(fileInfo.Name(), ".") {
				continue
			}
			file := templates.File{
				Path:  path.Join(filePath, fileInfo.Name()),
				Name:  fileInfo.Name(),
				Size:  formatFileSize(fileInfo.Size()),
				IsDir: fileInfo.IsDir(),
			}
			if file.IsDir {
				dirs = append(dirs, file)
			} else {
				files = append(files, file)
			}
		}
		page := templates.Page{
			Ok:       true,
			BackLink: filepath.Dir(filePath),
			Header:   filePath,
			Files:    append(dirs, files...),
			Version:  s.version,
		}
		if err = templates.ExecuteIndex(w, page); err != nil {
			s.sendError(w, err)
		}
	} else {
		s.sendFile(w, r, filePath)
	}
}

func (s *Server) sendErrorPage(w http.ResponseWriter, err error) {
	page := templates.Page{
		Ok:       false,
		BackLink: "/",
		Header:   err.Error(),
		Version:  s.version,
	}
	if err = templates.ExecuteIndex(w, page); err != nil {
		s.sendError(w, err)
	}
}

func (s *Server) sendFile(w http.ResponseWriter, r *http.Request, filePath string) {
	f, err := s.fs.Open(filePath)
	if err != nil {
		s.sendError(w, err)
		return
	}
	defer f.Close()
	if !s.rawEnabled {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		s.sendError(w, err)
		return
	}
	r.Header.Set("bytes-copied", formatFileSize(n))
}

func (s *Server) sendError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	s.errC <- err
}

func formatFileSize(size int64) string {
	var (
		unit   string
		factor int64
	)
	if factor = 1024 * 1024 * 1024; size >= factor {
		unit = "GB"
	} else if factor = 1024 * 1024; size >= factor {
		unit = "MB"
	} else if factor = 1024; size >= factor {
		unit = "KB"
	} else {
		unit = "B"
		factor = 1
	}
	return fmt.Sprintf("%0.2f%s", float64(size)/float64(factor), unit)
}
