package files

import (
	"io"
	"net/http"
)

type handler interface {
	parseUploadFile(r *http.Request) (io.Reader, string, error)
	handleDir(w io.Writer, dir string, entries []File) error
	handleError(w io.Writer, err error, code int) error
}
