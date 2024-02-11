package templates

import (
	_ "embed"
	"html/template"
	"io"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
)

type Page struct {
	Error       string
	Breadcrumbs []File
	Files       []File
	Version     string
}

type File struct {
	Path  string
	Name  string
	Size  string
	IsDir bool
}

func ExecuteIndex(w io.Writer, page Page) error {
	return indexTmpl.Execute(w, page)
}
