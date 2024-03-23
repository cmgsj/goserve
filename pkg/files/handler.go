package files

import "net/http"

type Handler interface {
	ContentType() string
	HandleDir(w http.ResponseWriter, dir string, entries []File) error
	HandleError(w http.ResponseWriter, err error, code int) error
}
