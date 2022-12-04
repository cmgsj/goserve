package handler

import (
	"errors"
	"fmt"
	"goserve/pkg/templates"
	"goserve/pkg/util"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func ServeFile(file string, fsize int64, download bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			err := templates.Index.Execute(w, templates.Page{
				Ok:       true,
				BackLink: "/",
				Header:   "",
				Files: []templates.File{
					{
						Path:  file,
						Name:  path.Base(file),
						Size:  util.GetFileSize(fsize),
						IsDir: false,
					},
				},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		SendFile(w, r, file, download)
	})
}

func ServeDir(dir string, download bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := dir + r.URL.Path
		if !util.IsValidPath(file) {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
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
			SendFile(w, r, file, download)
			return
		}
		entries, err := os.ReadDir(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var files []templates.File
		var dirs []templates.File
		for _, entry := range entries {
			size := " - "
			if info, err := entry.Info(); err == nil && !entry.IsDir() {
				size = util.GetFileSize(info.Size())
			}
			f := templates.File{
				Path:  path.Join(r.URL.Path, entry.Name()),
				Name:  entry.Name(),
				Size:  size,
				IsDir: entry.IsDir(),
			}
			if entry.IsDir() {
				dirs = append(dirs, f)
			} else {
				files = append(files, f)
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

func SendFile(w http.ResponseWriter, r *http.Request, filePath string, download bool) {
	if !util.IsValidPath(filePath) {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}
	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	if download {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.Header.Set("bytes-copied", util.GetFileSize(n))
}
