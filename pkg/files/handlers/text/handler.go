package text

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"text/tabwriter"

	"github.com/cmgsj/goserve/pkg/files"
)

func HandlerFactory() files.HandlerFactory {
	return func(s *files.Server) files.Handler {
		return (*handler)(s)
	}
}

type handler files.Server

func (h *handler) ContentType() string {
	return "text"
}

func (h *handler) HandleDir(w http.ResponseWriter, file string, entries []fs.DirEntry) error {
	var fileList []files.File

	if file != files.RootDir {
		fileList = append(fileList, files.File{
			Name:  files.ParentDir,
			IsDir: true,
		})
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
			Name:  info.Name(),
			IsDir: info.IsDir(),
		}

		if !f.IsDir {
			f.Size = files.FormatSize(info.Size())
		}

		fileList = append(fileList, f)
	}

	files.Sort(fileList)

	var buf bytes.Buffer

	for _, file := range fileList {
		buf.WriteString(file.Name)
		if file.IsDir {
			buf.WriteByte('/')
		} else {
			buf.WriteByte('\t')
			buf.WriteString(file.Size)
		}
		buf.WriteByte('\n')
	}

	tab := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)

	_, err := buf.WriteTo(tab)
	if err != nil {
		return err
	}

	return tab.Flush()
}

func (h *handler) HandleError(w http.ResponseWriter, err error, code int) error {
	_, err = fmt.Fprintf(w, "status: %s\nmessage: %s\n", http.StatusText(code), err.Error())
	return err
}
