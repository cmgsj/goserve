package templates

import (
	_ "embed"
	"html/template"
	"io"

	"github.com/cmgsj/goserve/pkg/files"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
)

type Index struct {
	Error       *Error
	Breadcrumbs []files.File
	Files       []files.File
	Version     string
}

type Error struct {
	Status  string
	Message string
}

func ExecuteIndex(w io.Writer, i Index) error {
	return indexTmpl.Execute(w, i)
}
