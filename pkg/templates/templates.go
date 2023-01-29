package templates

import (
	_ "embed"
	"html/template"
)

var (
	//go:embed index.html
	indexHTML string
	Index     = template.Must(template.New("index").Parse(indexHTML))
)

type Page struct {
	Ok       bool
	BackLink string
	Header   string
	Files    []File
	Version  string
}

type File struct {
	Path  string
	Name  string
	Size  string
	IsDir bool
}
