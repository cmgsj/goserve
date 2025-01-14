package files

import (
	_ "embed"
	"html/template"
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
	Data     *indexDataParams
	Error    *indexErrorParams
}

type indexDataParams struct {
	Breadcrumbs []File
	Files       []File
}

type indexErrorParams struct {
	Status  string
	Message string
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

func (h htmlHandler) handleDir(w http.ResponseWriter, r *http.Request, dir string, files []File) error {
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

	return h.handle(w, indexParams{
		Data: &indexDataParams{
			Breadcrumbs: breadcrumbs,
			Files:       files,
		},
	})
}

func (h htmlHandler) handleError(w http.ResponseWriter, r *http.Request, err error, code int) error {
	return h.handle(w, indexParams{
		Error: &indexErrorParams{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
	})
}

func (h htmlHandler) handle(w http.ResponseWriter, params indexParams) error {
	params.FilesURL = h.filesURL

	params.Uploads = h.uploads

	params.Version = h.version

	return indexTmpl.Execute(w, params)
}
