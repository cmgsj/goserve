package files

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"strings"
)

type Controller struct {
	fs      fs.FS
	exclude *regexp.Regexp
}

func NewController(fs fs.FS, exclude *regexp.Regexp) *Controller {
	return &Controller{
		fs:      fs,
		exclude: exclude,
	}
}

func (c *Controller) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func (c *Controller) FilesHTML() http.Handler {
	return c.files(newHTMLHandler())
}

func (c *Controller) FilesJSON() http.Handler {
	return c.files(newJSONHandler())
}

func (c *Controller) FilesText() http.Handler {
	return c.files(newTextHandler())
}

func (c *Controller) files(handler handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := r.PathValue("file")

		file = path.Clean(file)

		info, err := fs.Stat(c.fs, file)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		if !c.IsAllowed(file) {
			c.handleError(w, handler, newStaNotExistError(file), http.StatusNotFound)
			return
		}

		if !info.IsDir() {
			err = c.copyFile(w, file)
			if err != nil {
				c.handleError(w, handler, err, fsErrorStatusCode(err))
			}
			return
		}

		entries, err := c.readDir(file)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		err = handler.handleDir(w, file, entries)
		if err != nil {
			c.handleError(w, handler, err, http.StatusInternalServerError)
		}
	})
}

func (c *Controller) IsAllowed(file string) bool {
	if file == RootDir || c.exclude == nil {
		return true
	}

	for _, path := range strings.Split(file, "/") {
		if c.exclude.MatchString(path) {
			return false
		}
	}

	return true
}

func (c *Controller) copyFile(w io.Writer, file string) error {
	f, err := c.fs.Open(file)
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			slog.Error("failed to close file", "file", file, "error", err)
		}
	}()

	_, err = io.Copy(w, f)

	return err
}

func (c *Controller) readDir(dir string) ([]File, error) {
	entries, err := fs.ReadDir(c.fs, dir)
	if err != nil {
		return nil, err
	}

	var files []File

	if dir != RootDir {
		files = append(files, File{
			Path:  path.Dir(dir),
			Name:  ParentDir,
			IsDir: true,
		})
	}

	for _, entry := range entries {
		file := path.Join(dir, entry.Name())

		if !c.IsAllowed(file) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		files = append(files, File{
			Path:  file,
			Name:  info.Name(),
			Size:  FormatSize(info.Size()),
			IsDir: info.IsDir(),
		})
	}

	Sort(files)

	return files, nil
}

func (c *Controller) handleError(w http.ResponseWriter, handler handler, err error, code int) {
	slog.Error("an error ocurred", "error", err)

	w.WriteHeader(code)

	if handler != nil {
		handleErr := handler.handleError(w, err, code)
		if handleErr == nil {
			return
		}
		slog.Error("failed to handle error", "error", handleErr)
	}

	fmt.Fprintln(w, err.Error())
}
