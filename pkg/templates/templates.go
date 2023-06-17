package templates

import (
	_ "embed"
	"html/template"
	"io"

	"github.com/cmgsj/goserve/pkg/version"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
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

func ExecuteIndex(w io.Writer, p Page) error {
	p.Version = version.Version
	return indexTmpl.Execute(w, p)
}
