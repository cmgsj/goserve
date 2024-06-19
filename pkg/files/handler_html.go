package files

import (
	_ "embed"
	"html/template"
	"io"
	"net/http"
	"path"
	"strings"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
)

type indexParams struct {
	FilesURL string
	Uploads  bool
	Version  string
	Error    *indexErrorParams
	Data     *indexDataParams
}

type indexErrorParams struct {
	Status  string
	Message string
}

type indexDataParams struct {
	Breadcrumbs []File
	Files       []File
}

type htmlHandler struct {
	filesURL string
	uploads  bool
	version  string
}

func newHTMLHandler(filesURL string, uploads bool, version string) htmlHandler {
	return htmlHandler{
		filesURL: filesURL,
		uploads:  uploads,
		version:  version,
	}
}

func (h htmlHandler) handleDir(w io.Writer, dir string, files []File) error {
	var breadcrumbs []File

	if dir != RootDir {
		var pathPrefix string

		for _, name := range strings.Split(dir, "/") {
			pathPrefix = path.Join(pathPrefix, name)

			breadcrumbs = append(breadcrumbs, File{
				Path: pathPrefix,
				Name: name,
			})
		}
	}

	return indexTmpl.Execute(w, indexParams{
		FilesURL: h.filesURL,
		Uploads:  h.uploads,
		Version:  h.version,
		Data: &indexDataParams{
			Breadcrumbs: breadcrumbs,
			Files:       files,
		},
	})
}

func (h htmlHandler) handleError(w io.Writer, err error, code int) error {
	return indexTmpl.Execute(w, indexParams{
		FilesURL: h.filesURL,
		Uploads:  h.uploads,
		Version:  h.version,
		Error: &indexErrorParams{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
	})
}
