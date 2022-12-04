package handler

import (
	"errors"
	"fmt"
	"goserve/pkg/templates"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func ServeFile(file string, fsize int64, serveAsText bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		if r.URL.Path != "/" {
			err = sendFile(w, r, file, serveAsText)
		} else {
			err = templates.Index.Execute(w, templates.Page{
				Ok:       true,
				BackLink: "/",
				Header:   "",
				Files: []templates.File{
					{
						Path:  file,
						Name:  path.Base(file),
						Size:  FormatFileSize(fsize),
						IsDir: false,
					},
				}})
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func ServeDir(dir string, serveAsText bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := ValidateUrlPath(r.URL.Path)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		file := dir + r.URL.Path
		fstat, err := os.Stat(file)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = templates.Index.Execute(w, templates.Page{
					Ok:       false,
					BackLink: "/",
					Header:   fmt.Sprintf("file not found: %s", file),
				})
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if !fstat.IsDir() {
			if err = sendFile(w, r, file, serveAsText); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		entries, err := os.ReadDir(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var files, dirs []templates.File
		for _, entry := range entries {
			size := " - "
			if info, err := entry.Info(); err == nil && !entry.IsDir() {
				size = FormatFileSize(info.Size())
			}
			fileTmpl := templates.File{
				Path:  path.Join(r.URL.Path, entry.Name()),
				Name:  entry.Name(),
				Size:  size,
				IsDir: entry.IsDir(),
			}
			if entry.IsDir() {
				dirs = append(dirs, fileTmpl)
			} else {
				files = append(files, fileTmpl)
			}
		}
		err = templates.Index.Execute(w, templates.Page{
			Ok:       true,
			BackLink: path.Dir(strings.TrimRight(r.URL.Path, "/")),
			Header:   strings.Trim(r.URL.Path, "/"),
			Files:    append(dirs, files...),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func sendFile(w http.ResponseWriter, r *http.Request, filePath string, serveAsText bool) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if !serveAsText {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		return err
	}
	r.Header.Set("bytes-copied", FormatFileSize(n))
	return nil
}

func ValidateUrlPath(p string) error {
	if strings.Contains(p, "..") || strings.Contains(p, "~") {
		return fmt.Errorf("invalid path %s: must not contain '..' or '~'", p)
	}
	return nil
}

func FormatFileSize(numBytes int64) string {
	var unit string
	var conv int64
	if conv = 1024 * 1024 * 2014; numBytes > conv {
		unit = "GB"
	} else if conv = 1024 * 1024; numBytes > conv {
		unit = "MB"
	} else if conv = 1024; numBytes > conv {
		unit = "KB"
	} else {
		unit = "B"
		conv = 1
	}
	return fmt.Sprintf("%.2f%s", float64(numBytes)/float64(conv), unit)
}
