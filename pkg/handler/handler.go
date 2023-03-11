package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cmgsj/goserve/pkg/templates"
)

func ServeFile(rootFile string, skipDotFiles, rawEnabled bool, version string, errCh chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			filePath = path.Clean(r.URL.Path)
			fullPath = path.Join(rootFile, filePath)
		)
		info, err := os.Stat(fullPath)
		if err != nil {
			sendErrorPage(w, err, version, errCh)
			return
		}
		if info.IsDir() {
			entries, err := os.ReadDir(fullPath)
			if err != nil {
				sendErrorPage(w, err, version, errCh)
				return
			}
			var (
				dirs  = make([]*templates.File, 0, len(entries))
				files = make([]*templates.File, 0)
			)
			for _, entry := range entries {
				if skipDotFiles && strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				info, err = entry.Info()
				if err != nil {
					errCh <- err
					continue
				}
				file := &templates.File{
					Path:  path.Join(filePath, info.Name()),
					Name:  info.Name(),
					Size:  formatFileSize(info.Size()),
					IsDir: info.IsDir(),
				}
				if file.IsDir {
					dirs = append(dirs, file)
				} else {
					files = append(files, file)
				}
			}
			page := &templates.Page{
				Ok:       true,
				BackLink: filepath.Dir(filePath),
				Header:   filePath,
				Files:    append(dirs, files...),
				Version:  version,
			}
			if err = templates.ExecuteIndex(w, page); err != nil {
				errCh <- err
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if err = sendFile(w, r, fullPath, rawEnabled); err != nil {
			errCh <- err
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func sendErrorPage(w http.ResponseWriter, err error, version string, errCh chan<- error) {
	page := &templates.Page{
		Ok:       false,
		BackLink: "/",
		Header:   err.Error(),
		Version:  version,
	}
	if err = templates.ExecuteIndex(w, page); err != nil {
		errCh <- err
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendFile(w http.ResponseWriter, r *http.Request, filePath string, rawEnabled bool) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if !rawEnabled {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		return err
	}
	r.Header.Set("bytes-copied", formatFileSize(n))
	return nil
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
