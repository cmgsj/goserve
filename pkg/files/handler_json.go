package files

import (
	"encoding/json"
	"io"
	"net/http"
)

type jsonHandler struct {
	indent bool
}

func newJSONHandler(indent bool) jsonHandler {
	return jsonHandler{
		indent: indent,
	}
}

func (h jsonHandler) handleDir(w io.Writer, dir string, entries []File) error {
	return h.encode(w, entries)
}

func (h jsonHandler) handleError(w io.Writer, err error, code int) error {
	return h.encode(w, map[string]interface{}{
		"status":  http.StatusText(code),
		"message": err.Error(),
	})
}

func (h jsonHandler) encode(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)

	if h.indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(v)
}
