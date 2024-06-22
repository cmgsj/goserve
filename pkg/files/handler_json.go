package files

import (
	"encoding/json"
	"net/http"
)

type jsonHandler struct {
}

func newJSONHandler() jsonHandler {
	return jsonHandler{}
}

func (h jsonHandler) handleDir(w http.ResponseWriter, r *http.Request, dir string, files []File) error {
	return h.encode(w, r, files)
}

func (h jsonHandler) handleError(w http.ResponseWriter, r *http.Request, err error, code int) error {
	return h.encode(w, r, map[string]interface{}{
		"status":  http.StatusText(code),
		"message": err.Error(),
	})
}

func (h jsonHandler) encode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	compact := r.URL.Query().Has("compact")

	encoder := json.NewEncoder(w)

	if !compact {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(v)
}
