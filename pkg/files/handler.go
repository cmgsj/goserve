package files

import "net/http"

type handler interface {
	handleDir(w http.ResponseWriter, r *http.Request, dir string, files []File) error
	handleError(w http.ResponseWriter, r *http.Request, err error, code int) error
}
