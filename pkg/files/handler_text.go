package files

import (
	"bytes"
	"fmt"
	"net/http"
	"text/tabwriter"
)

type textHandler struct{}

func newTextHandler() textHandler {
	return textHandler{}
}

func (h textHandler) handleDir(w http.ResponseWriter, r *http.Request, dir string, files []File) error {
	fullpath := parseQueryBool(r.URL, "fullpath")

	var buf bytes.Buffer

	for _, file := range files {
		if fullpath {
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

func (h textHandler) handleError(w http.ResponseWriter, r *http.Request, err error, code int) error {
	_, err = fmt.Fprintf(w, "%s\n\n%s\n", http.StatusText(code), err.Error())
	return err
}
