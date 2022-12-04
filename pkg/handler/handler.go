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

func ServeFile(file string, fsize int64, plaintext bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			if err := sendFile(w, r, file, plaintext); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		err := templates.Index.Execute(w, templates.Page{
			Ok:       true,
			BackLink: "/",
			Header:   "",
			Files: []templates.File{
				{
					Path:  file,
					Name:  path.Base(file),
					Size:  util.FormatFileSize(fsize),
					IsDir: false,
				},
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func ServeDir(dir string, plaintext bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := dir + r.URL.Path
		err := util.ValidatePath(file)
		if err != nil {
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
			if err = sendFile(w, r, file, plaintext); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
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
				size = util.FormatFileSize(info.Size())
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

func sendFile(w http.ResponseWriter, r *http.Request, filePath string, plaintext bool) error {
	err := util.ValidatePath(filePath)
	if err != nil {
		return err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if !plaintext {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", path.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		return err
	}
	r.Header.Set("bytes-copied", util.FormatFileSize(n))
	return nil
}
