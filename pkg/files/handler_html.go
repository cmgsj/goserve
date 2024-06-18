package files

import (
	_ "embed"
	"io"
	"net/http"
	"path"
	"strings"
	"text/template"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
)

type indexParams struct {
	Error       *errorParams
	Breadcrumbs []File
	Files       []File
	Uploads     bool
	Version     string
}

type errorParams struct {
	Status  string
	Message string
}

type htmlHandler struct {
	uploads bool
	version string
}

func newHTMLHandler(uploads bool, version string) htmlHandler {
	return htmlHandler{
		uploads: uploads,
		version: version,
	}
}

func (h htmlHandler) handleDir(w io.Writer, dir string, files []File) error {
	var breadcrumbs []File

	if dir != RootDir {
		var prefix string

		for _, name := range strings.Split(dir, "/") {
			prefix = path.Join(prefix, name)

			breadcrumbs = append(breadcrumbs, File{
				Path: prefix,
				Name: name,
			})
		}
	}

	return indexTmpl.Execute(w, indexParams{
		Breadcrumbs: breadcrumbs,
		Files:       files,
		Uploads:     h.uploads,
		Version:     h.version,
	})
}

func (h htmlHandler) handleError(w io.Writer, err error, code int) error {
	return indexTmpl.Execute(w, indexParams{
		Error: &errorParams{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
		Version: h.version,
	})
}
