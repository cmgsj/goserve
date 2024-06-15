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

type indexData struct {
	Error       *errorData
	Breadcrumbs []File
	Upload      bool
	Files       []File
	Version     string
}

type errorData struct {
	Status  string
	Message string
}

type htmlHandler struct {
	upload  bool
	version string
}

func newHTMLHandler(upload bool, version string) htmlHandler {
	return htmlHandler{
		upload:  upload,
		version: version,
	}
}

func (h htmlHandler) handleDir(w io.Writer, dir string, entries []File) error {
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

	return indexTmpl.Execute(w, indexData{
		Breadcrumbs: breadcrumbs,
		Upload:      h.upload,
		Files:       entries,
		Version:     h.version,
	})
}

func (h htmlHandler) handleError(w io.Writer, err error, code int) error {
	return indexTmpl.Execute(w, indexData{
		Error: &errorData{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
		Version: h.version,
	})
}
