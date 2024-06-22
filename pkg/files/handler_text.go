package files

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"text/tabwriter"
)

type textHandler struct {
	fullpath bool
}

func newTextHandler(fullpath bool) textHandler {
	return textHandler{
		fullpath: fullpath,
	}
}

func (h textHandler) handleDir(w io.Writer, dir string, files []File) error {
	var buf bytes.Buffer

	for _, file := range files {
		if h.fullpath {
			buf.WriteString(file.Path)
		} else {
			buf.WriteString(file.Name)
		}
		if file.IsDir {
			buf.WriteByte('/')
		} else {
			buf.WriteByte('\t')
			buf.WriteString(file.Size)
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

func (h textHandler) handleError(w io.Writer, err error, code int) error {
	_, err = fmt.Fprintf(w, "%s\n\n%s\n", http.StatusText(code), err.Error())
	return err
}
