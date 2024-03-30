package files

import "io"

type Handler interface {
	ContentType() string
	HandleDir(w io.Writer, dir string, entries []File) error
	HandleError(w io.Writer, err error, code int) error
}
