package handler

import (
	"fmt"
	"goserve/pkg/file"
	"goserve/pkg/middleware"
	"goserve/pkg/templates"
	"io"
	"net/http"
	"os"
	"path"
)

func ServeRoot(root *file.Entry, serveAsText bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := root.FindMatch(r.URL.Path)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			err = templates.Index.Execute(w, templates.Page{
				Ok:       false,
				BackLink: "/",
				Header:   err.Error(),
			})
		} else if f.IsDir {
			var fileTmpls []templates.File
			for _, child := range f.Children {
				fileTmpls = append(fileTmpls, templates.File{
					Path:  fmt.Sprintf("/%s", child.Path),
					Name:  child.Name,
					Size:  child.Size,
					IsDir: child.IsDir,
				})
			}
			err = templates.Index.Execute(w, templates.Page{
				Ok:       true,
				BackLink: fmt.Sprintf("/%s", path.Dir(f.Path)),
				Header:   f.Path,
				Files:    fileTmpls,
			})
		} else {
			err = sendFile(w, r, f.Path, serveAsText)
		}
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
	r.Header.Set(middleware.BytesCopiedHeader, file.FormatSize(n))
	return nil
}
