package files

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"
)

type Controller struct {
	fs           fs.FS
	exclude      *regexp.Regexp
	handlers     map[string]Handler
	contentTypes []string
}

func NewController(fs fs.FS, exclude *regexp.Regexp, handlers ...Handler) *Controller {
	controller := &Controller{
		fs:       fs,
		exclude:  exclude,
		handlers: make(map[string]Handler),
	}

	for _, handler := range handlers {
		controller.handlers[handler.ContentType()] = handler
		controller.contentTypes = append(controller.contentTypes, handler.ContentType())
	}

	slices.Sort(controller.contentTypes)

	return controller
}

func (c *Controller) FilesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.PathValue("content_type")
		file := r.PathValue("file")

		handler, ok := c.handlers[contentType]
		if !ok {
			c.handleError(w, nil, newUnsupportedContentTypeError(contentType, c.contentTypes), http.StatusBadRequest)
			return
		}

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

		err = handler.HandleDir(w, file, entries)
		if err != nil {
			c.handleError(w, handler, err, http.StatusInternalServerError)
		}
	})
}

func (c *Controller) ContentTypesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, c.contentTypes)
	})
}

func (c *Controller) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func (c *Controller) ContentTypes() []string {
	return c.contentTypes
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

func (c *Controller) copyFile(w http.ResponseWriter, file string) error {
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

func (c *Controller) handleError(w http.ResponseWriter, handler Handler, err error, code int) {
	slog.Error("an error ocurred", "error", err)

	w.WriteHeader(code)

	if handler != nil {
		handleErr := handler.HandleError(w, err, code)
		if handleErr == nil {
			return
		}
		slog.Error("failed to handle error", "error", handleErr)
	}

	fmt.Fprintln(w, err.Error())
}
