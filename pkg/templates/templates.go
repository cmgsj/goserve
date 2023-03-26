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

type (
	Page struct {
		Ok       bool
		BackLink string
		Header   string
		Files    []File
		Version  string
	}
	File struct {
		Path  string
		Name  string
		Size  string
		IsDir bool
	}
)

func ExecuteIndex(w io.Writer, page Page) error {
	return indexTmpl.Execute(w, page)
}
