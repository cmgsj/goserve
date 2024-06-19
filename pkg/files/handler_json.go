package files

import (
	"encoding/json"
	"io"
	"net/http"
)

type jsonHandler struct {
	compact bool
}

func newJSONHandler(compact bool) jsonHandler {
	return jsonHandler{
		compact: compact,
	}
}

func (h jsonHandler) handleDir(w io.Writer, dir string, files []File) error {
	return h.encode(w, files)
}

func (h jsonHandler) handleError(w io.Writer, err error, code int) error {
	return h.encode(w, map[string]interface{}{
		"status":  http.StatusText(code),
		"message": err.Error(),
	})
}

func (h jsonHandler) encode(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)

	if !h.compact {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(v)
}
