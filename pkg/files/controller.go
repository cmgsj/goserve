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

func (c *Controller) UploadHTML(uploadDir, redirectURL string) http.Handler {
	return c.upload(newHTMLHandler(), uploadDir, redirectURL)
}

func (c *Controller) UploadJSON(uploadDir, redirectURL string) http.Handler {
	return c.upload(newJSONHandler(), uploadDir, redirectURL)
}

func (c *Controller) UploadText(uploadDir, redirectURL string) http.Handler {
	return c.upload(newTextHandler(), uploadDir, redirectURL)
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

func (c *Controller) upload(handler handler, uploadDir, redirectURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, filename, err := handler.parseUploadFile(r)
		if err != nil {
			c.handleError(w, handler, err, http.StatusBadRequest)
			return
		}

		path := filepath.Join(uploadDir, fmt.Sprintf("%d_%s", time.Now().Unix(), filename))

		_, err = os.Stat(path)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				c.handleError(w, handler, err, fsErrorStatusCode(err))
				return
			}
		} else {
			c.handleError(w, handler, fs.ErrExist, http.StatusBadRequest)
			return
		}

		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		defer func() {
			err := f.Close()
			if err != nil {
				slog.Error("failed to close uploaded file", "path", path, "error", err)
			}
		}()

		_, err = io.Copy(f, file)
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		err = f.Sync()
		if err != nil {
			c.handleError(w, handler, err, fsErrorStatusCode(err))
			return
		}

		if redirectURL != "" {
			http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
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
		err := f.Close()
		if err != nil {
			slog.Error("failed to close copied file", "path", file, "error", err)
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

	handleErr := handler.handleError(w, err, code)
	if handleErr == nil {
		return
	}
	slog.Error("failed to handle error", "error", handleErr)

	fmt.Fprintln(w, err.Error())
}
