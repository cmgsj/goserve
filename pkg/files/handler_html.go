package files

import (
	_ "embed"
	"io"
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/cmgsj/goserve/internal/version"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
)

type indexData struct {
	Error       *errorData
	Breadcrumbs []File
	Files       []File
	Version     string
}

type errorData struct {
	Status  string
	Message string
}

type htmlHandler struct{}

func newHTMLHandler() htmlHandler {
	return htmlHandler{}
}

func (h htmlHandler) parseUploadFile(r *http.Request) (io.Reader, string, error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, "", err
	}

	return file, header.Filename, nil
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
		Files:       entries,
		Version:     version.String(),
	})
}

func (h htmlHandler) handleError(w io.Writer, err error, code int) error {
	return indexTmpl.Execute(w, indexData{
		Error: &errorData{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
		Version: version.String(),
	})
}
