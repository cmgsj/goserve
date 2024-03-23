package json

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"path"

	"github.com/cmgsj/goserve/pkg/files"
)

func HandlerFactory() files.HandlerFactory {
	return func(s *files.Server) files.Handler {
		return (*handler)(s)
	}
}

type handler files.Server

func (h *handler) ContentType() string {
	return "json"
}

func (h *handler) SendDir(w http.ResponseWriter, file string, entries []fs.DirEntry) error {
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

	encoder := json.NewEncoder(w)

	encoder.SetIndent("", "  ")

	return encoder.Encode(fileList)
}

func (h *handler) SendError(w http.ResponseWriter, err error, code int) {
	(*files.Server)(h).SendError(w, err, code)
}
