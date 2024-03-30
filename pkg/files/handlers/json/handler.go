package json

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/cmgsj/goserve/pkg/files"
)

var _ files.Handler = (*Handler)(nil)

type Handler struct {
	indent bool
}

func NewHandler(indent bool) *Handler {
	return &Handler{
		indent: indent,
	}
}

func (h *Handler) ContentType() string {
	return "json"
}

func (h *Handler) HandleDir(w io.Writer, dir string, entries []files.File) error {
	return h.encode(w, entries)
}

func (h *Handler) HandleError(w io.Writer, err error, code int) error {
	return h.encode(w, map[string]interface{}{
		"status":  http.StatusText(code),
		"message": err.Error(),
	})
}

func (h *Handler) encode(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)

	if h.indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(v)
}
