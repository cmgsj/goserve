package handler

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

type FileServerConfig struct {
	FS           afero.Fs
	SkipDotFiles bool
	RawEnabled   bool
	Version      string
	ErrC         chan<- error
}

func FileServer(config FileServerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := path.Clean(r.URL.Path)
		info, err := config.FS.Stat(filePath)
		if err != nil {
			sendErrorPage(w, err, config.Version, config.ErrC)
			return
		}
		if info.IsDir() {
			filesInfo, err := afero.ReadDir(config.FS, filePath)
			if err != nil {
				sendErrorPage(w, err, config.Version, config.ErrC)
				return
			}
			var (
				dirs  = make([]templates.File, 0, len(filesInfo))
				files = make([]templates.File, 0)
			)
			for _, fileInfo := range filesInfo {
				if config.SkipDotFiles && strings.HasPrefix(fileInfo.Name(), ".") {
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
				Version:  config.Version,
			}
			if err = templates.ExecuteIndex(w, page); err != nil {
				sendError(w, err, config.ErrC)
			}
		} else {
			sendFile(w, r, config.FS, filePath, config.RawEnabled, config.ErrC)
		}
	})
}

func sendErrorPage(w http.ResponseWriter, err error, version string, errC chan<- error) {
	page := templates.Page{
		Ok:       false,
		BackLink: "/",
		Header:   err.Error(),
		Version:  version,
	}
	if err = templates.ExecuteIndex(w, page); err != nil {
		sendError(w, err, errC)
	}
}

func sendFile(w http.ResponseWriter, r *http.Request, fsys afero.Fs, filePath string, rawEnabled bool, errC chan<- error) {
	f, err := fsys.Open(filePath)
	if err != nil {
		sendError(w, err, errC)
		return
	}
	defer f.Close()
	if !rawEnabled {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		sendError(w, err, errC)
		return
	}
	r.Header.Set("bytes-copied", formatFileSize(n))
}

func sendError(w http.ResponseWriter, err error, errC chan<- error) {
	errC <- err
	http.Error(w, err.Error(), http.StatusInternalServerError)
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
