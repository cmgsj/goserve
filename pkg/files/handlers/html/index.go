package html

import (
	_ "embed"
	"html/template"

	"github.com/cmgsj/goserve/pkg/files"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
)

type indexData struct {
	Error       *errorData
	Breadcrumbs []files.File
	Files       []files.File
	Version     string
}

type errorData struct {
	Status  string
	Message string
}
