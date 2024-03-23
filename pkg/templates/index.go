package templates

import (
	"cmp"
	_ "embed"
	"html/template"
	"io"
	"slices"
)

var (
	//go:embed index.html
	indexHTML string
	indexTmpl = template.Must(template.New("index").Parse(indexHTML))
)

type Index struct {
	Error       *Error
	Breadcrumbs []File
	Files       []File
	Version     string
}

type Error struct {
	Status  string
	Message string
}

type File struct {
	Path  string
	Name  string
	Size  string
	IsDir bool
}

func ExecuteIndex(w io.Writer, i Index) error {
	return indexTmpl.Execute(w, i)
}

func SortFiles(files []File) {
	slices.SortFunc(files, func(x, y File) int {
		if x.IsDir != y.IsDir {
			if x.IsDir {
				return -1
			}
			return +1
		}
		return cmp.Compare(x.Name, y.Name)
	})
}
