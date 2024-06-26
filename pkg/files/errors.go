package files

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
)

func fsNotExistError(filePath string) error {
	return fmt.Errorf("stat %s: no such file or directory", filePath)
}

func fsErrorStatusCode(err error) int {
	switch {
	case err == nil:
		return http.StatusOK

	case errors.Is(err, fs.ErrInvalid):
		return http.StatusBadRequest

	case errors.Is(err, fs.ErrPermission):
		return http.StatusForbidden

	case errors.Is(err, fs.ErrNotExist):
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}
