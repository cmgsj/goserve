package templates

import (
	_ "embed"
	"html/template"
	"io"
	"sort"
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
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})
}
