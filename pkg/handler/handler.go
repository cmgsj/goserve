package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cmgsj/goserve/pkg/file"
	"github.com/cmgsj/goserve/pkg/format"
	"github.com/cmgsj/goserve/pkg/templates"
)

func ServeFileTree(root *file.FileTree, rawEnabled bool, version string, errch chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := root.FindMatch(r.URL.Path)
		if err != nil {
			if errors.Is(err, file.ErrFileNotFound) {
				w.WriteHeader(http.StatusNotFound)
				err = templates.ExecuteIndex(w, templates.Page{
					Ok:       false,
					BackLink: "/",
					Header:   err.Error(),
					Version:  version,
				})
			}
		} else if f.IsDir {
			var files, dirs []templates.File
			for _, child := range f.Children {
				if child.IsBroken {
					continue
				}
				file := templates.File{
					Path:  strings.TrimPrefix(child.Path, root.Path),
					Name:  child.Name,
					Size:  format.FileSize(child.Size),
					IsDir: child.IsDir,
				}
				if child.IsDir {
					dirs = append(dirs, file)
				} else {
					files = append(files, file)
				}
			}
			err = templates.ExecuteIndex(w, templates.Page{
				Ok:       true,
				BackLink: filepath.Dir(strings.TrimPrefix(f.Path, root.Path)),
				Header:   "/" + strings.TrimPrefix(strings.TrimPrefix(f.Path, root.Path), "/"),
				Files:    append(dirs, files...),
				Version:  version,
			})
		} else {
			err = sendFile(w, r, f.Path, rawEnabled)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			errch <- err
		}
	})
}

func sendFile(w http.ResponseWriter, r *http.Request, filePath string, rawEnabled bool) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if !rawEnabled {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filePath)))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	n, err := io.Copy(w, f)
	if err != nil {
		return err
	}
	r.Header.Set("bytes-copied", format.FileSize(n))
	return nil
}
