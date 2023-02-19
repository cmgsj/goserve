package templates

import (
	_ "embed"
	"errors"
	"html/template"
	"io"
)

var (
	//go:embed index.html
	indexHtml  string
	indexTmpl  = template.Must(template.New("index").Parse(indexHtml))
	ErrNilPage = errors.New("nil page")
)

type Page struct {
	Ok       bool
	BackLink string
	Header   string
	Files    []*File
	Version  string
}

type File struct {
	Path  string
	Name  string
	Size  string
	IsDir bool
}

func ExecuteIndex(w io.Writer, page *Page) error {
	if page == nil {
		return ErrNilPage
	}
	return indexTmpl.Execute(w, page)
}
