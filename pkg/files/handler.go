package files

import (
	"io/fs"
	"net/http"
)

type Handler interface {
	ContentType() string
	HandleDir(w http.ResponseWriter, file string, entries []fs.DirEntry) error
	HandleError(w http.ResponseWriter, err error, code int) error
}

type HandlerFactory func(s *Server) Handler
