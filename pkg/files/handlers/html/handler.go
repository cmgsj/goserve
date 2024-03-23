package html

import (
	"net/http"
	"path"
	"strings"

	"github.com/cmgsj/goserve/pkg/files"
)

type Handler struct {
	version string
}

func NewHandler(version string) *Handler {
	return &Handler{
		version: version,
	}
}

func (h *Handler) ContentType() string {
	return "html"
}

func (h *Handler) HandleDir(w http.ResponseWriter, dir string, entries []files.File) error {
	var breadcrumbs []files.File

	if dir != files.RootDir {
		var prefix string

		for _, name := range strings.Split(dir, "/") {
			prefix = path.Join(prefix, name)

			breadcrumbs = append(breadcrumbs, files.File{
				Path: prefix,
				Name: name,
			})
		}
	}

	return indexTmpl.Execute(w, indexData{
		Breadcrumbs: breadcrumbs,
		Files:       entries,
		Version:     h.version,
	})
}

func (h *Handler) HandleError(w http.ResponseWriter, err error, code int) error {
	return indexTmpl.Execute(w, indexData{
		Error: &errorData{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
		Version: h.version,
	})
}
