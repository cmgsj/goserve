package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Controller struct {
	fileSystem  fs.FS
	htmlHandler handler
	jsonHandler handler
	textHandler handler
	config      ControllerConfig
}

type ControllerConfig struct {
	FilesURL         string
	ExcludePattern   *regexp.Regexp
	Uploads          bool
	UploadsDir       string
	UploadsTimestamp bool
	CompactJSON      bool
	Version          string
}

func NewController(fileSystem fs.FS, config ControllerConfig) *Controller {
	if config.FilesURL == "" {
		config.FilesURL = "/"
	}

	return &Controller{
		fileSystem:  fileSystem,
		htmlHandler: newHTMLHandler(config.FilesURL, config.Uploads, config.Version),
		jsonHandler: newJSONHandler(config.CompactJSON),
		textHandler: newTextHandler(),
		config:      config,
	}
}

func (c *Controller) ListFiles() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := c.requestHandler(r)

		filePath := r.PathValue("file")

		filePath = path.Clean(filePath)

		if c.isForbidden(filePath) {
			c.handleError(w, handler, fsNotExistError(filePath), http.StatusNotFound)
			return
		}

		fileInfo, err := fs.Stat(c.fileSystem, filePath)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		if !fileInfo.IsDir() {
			err = c.copyFile(w, filePath)
			if err != nil {
				c.handleError(w, handler, err, fsErrorStatusCode(err))
			}
			return
		}

		files, err := c.readDir(filePath)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		err = handler.handleDir(w, filePath, files)
		if err != nil {
			c.handleError(w, handler, err, http.StatusInternalServerError)
		}
	})
}

func (c *Controller) UploadFile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := c.requestHandler(r)

		if !c.config.Uploads {
			c.handleError(w, handler, fs.ErrPermission, http.StatusForbidden)
			return
		}

		formFile, header, err := r.FormFile("file")
		if err != nil {
			c.handleError(w, handler, err, http.StatusBadRequest)
			return
		}

		if c.config.UploadsTimestamp {
			header.Filename = time.Now().UTC().Format(time.DateTime) + " " + header.Filename
		}

		filePath := filepath.Join(c.config.UploadsDir, header.Filename)

		_, err = os.Stat(filePath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				c.handleError(w, handler, err, fsErrorStatusCode(err))
				return
			}
		} else {
			c.handleError(w, handler, fs.ErrExist, http.StatusBadRequest)
			return
		}

		osFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		defer func() {
			err := osFile.Close()
			if err != nil {
				slog.Error("failed to close uploaded file", "path", filePath, "error", err)
			}
		}()

		_, err = io.Copy(osFile, formFile)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		err = osFile.Sync()
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		http.Redirect(w, r, c.config.FilesURL, http.StatusFound)
	})
}

func (c *Controller) requestHandler(r *http.Request) handler {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return c.jsonHandler

	case "text/plain":
		return c.textHandler

	default:
		return c.htmlHandler
	}
}

func (c *Controller) isForbidden(filePath string) bool {
	if filePath == RootDir {
		return false
	}

	if c.config.ExcludePattern != nil {
		for _, part := range strings.Split(filePath, "/") {
			if c.config.ExcludePattern.MatchString(part) {
				return true
			}
		}
	}

	return false
}

func (c *Controller) copyFile(dst io.Writer, filePath string) error {
	fsFile, err := c.fileSystem.Open(filePath)
	if err != nil {
		return err
	}

	defer func() {
		err := fsFile.Close()
		if err != nil {
			slog.Error("failed to close copied file", "path", filePath, "error", err)
		}
	}()

	_, err = io.Copy(dst, fsFile)
	return err
}

func (c *Controller) readDir(filePath string) ([]File, error) {
	entries, err := fs.ReadDir(c.fileSystem, filePath)
	if err != nil {
		return nil, err
	}

	var files []File

	if filePath != RootDir {
		files = append(files, File{
			Path:  path.Dir(filePath),
			Name:  ParentDir,
			IsDir: true,
		})
	}

	for _, entry := range entries {
		entryPath := path.Join(filePath, entry.Name())

		if c.isForbidden(entryPath) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		file := File{
			Path:  entryPath,
			Name:  info.Name(),
			IsDir: info.IsDir(),
		}

		if !file.IsDir {
			file.Size = FormatSize(info.Size(), 2)
		}

		files = append(files, file)
	}

	Sort(files)

	return files, nil
}

func (c *Controller) handleError(w http.ResponseWriter, handler handler, err error, code int) {
	slog.Error("an error ocurred", "error", err)

	w.WriteHeader(code)

	herr := handler.handleError(w, err, code)

	if herr != nil {
		slog.Error("failed to handle error", "error", herr)

		fmt.Fprintln(w, err.Error())
	}
}
