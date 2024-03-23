package files

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

func newUnsupportedContentTypeError(contentType string, contentTypes []string) error {
	return fmt.Errorf("unsupported content type %q, supported: [%s]", contentType, strings.Join(contentTypes, ","))
}

func newFileNotFoundError(file string) error {
	return fmt.Errorf("stat %s: no such file or directory", file)
}

func fsErrorStatusCode(err error) int {
	switch {
	case errors.Is(err, fs.ErrInvalid):
		return http.StatusBadRequest

	case errors.Is(err, fs.ErrPermission):
		return http.StatusUnauthorized

	case errors.Is(err, fs.ErrNotExist):
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}
