package text

import (
	"bytes"
	"fmt"
	"net/http"
	"text/tabwriter"

	"github.com/cmgsj/goserve/pkg/files"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ContentType() string {
	return "text"
}

func (h *Handler) HandleDir(w http.ResponseWriter, dir string, entries []files.File) error {
	var buf bytes.Buffer

	for _, entry := range entries {
		buf.WriteString(entry.Name)
		if entry.IsDir {
			buf.WriteByte('/')
		} else {
			buf.WriteByte('\t')
			buf.WriteString(entry.Size)
		}
		buf.WriteByte('\n')
	}

	tab := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)

	_, err := buf.WriteTo(tab)
	if err != nil {
		return err
	}

	return tab.Flush()
}

func (h *Handler) HandleError(w http.ResponseWriter, err error, code int) error {
	_, err = fmt.Fprintf(w, "status: %s\nmessage: %s\n", http.StatusText(code), err.Error())
	return err
}
