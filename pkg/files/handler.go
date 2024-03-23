package files

import (
	"io/fs"
	"net/http"
)

type Handler interface {
	ContentType() string
	SendDir(w http.ResponseWriter, file string, entries []fs.DirEntry) error
	SendError(w http.ResponseWriter, err error, code int)
}

type HandlerFactory func(s *Server) Handler
