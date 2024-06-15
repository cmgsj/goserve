package files

import "io"

type handler interface {
	handleDir(w io.Writer, dir string, entries []File) error
	handleError(w io.Writer, err error, code int) error
}
