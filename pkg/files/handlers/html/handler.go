package html

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/cmgsj/goserve/pkg/files"
)

func HandlerFactory() files.HandlerFactory {
	return func(s *files.Server) files.Handler {
		return (*handler)(s)
	}
}

type handler files.Server

func (h *handler) ContentType() string {
	return "html"
}

func (h *handler) HandleDir(w http.ResponseWriter, file string, entries []fs.DirEntry) error {
	var breadcrumbList, fileList []files.File

	if file != files.RootDir {
		fileList = append(fileList, files.File{
			Path:  path.Dir(file),
			Name:  files.ParentDir,
			IsDir: true,
		})

		var prefix string

		for _, name := range strings.Split(file, "/") {
			prefix = path.Join(prefix, name)

			breadcrumbList = append(breadcrumbList, files.File{
				Path: prefix,
				Name: name,
			})
		}
	}

	for _, entry := range entries {
		entryPath := path.Join(file, entry.Name())

		if !(*files.Server)(h).IsAllowed(entryPath) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		f := files.File{
			Path:  entryPath,
			Name:  info.Name(),
			IsDir: info.IsDir(),
		}

		if !f.IsDir {
			f.Size = files.FormatSize(info.Size())
		}

		fileList = append(fileList, f)
	}

	files.Sort(fileList)

	return indexTmpl.Execute(w, indexData{
		Breadcrumbs: breadcrumbList,
		Files:       fileList,
		Version:     (*files.Server)(h).Version(),
	})
}

func (h *handler) HandleError(w http.ResponseWriter, err error, code int) error {
	return indexTmpl.Execute(w, indexData{
		Error: &errorData{
			Status:  http.StatusText(code),
			Message: err.Error(),
		},
		Version: (*files.Server)(h).Version(),
	})
}
