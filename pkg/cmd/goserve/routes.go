package goserve

import (
	"bytes"
	"net/http"
	"os"
	"text/tabwriter"
)

type route struct {
	patterns    []string
	description string
	handler     http.Handler
	disabled    bool
}

func registerRoutes(mux *http.ServeMux, routes []route) error {
	var buf bytes.Buffer

	for _, route := range routes {
		if !route.disabled {
			for _, pattern := range route.patterns {
				mux.Handle(pattern, route.handler)
				buf.WriteString(sprintfln("  %s\t->\t%s", pattern, route.description))
			}
		}
	}

	tab := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	_, err := buf.WriteTo(tab)
	if err != nil {
		return err
	}

	return tab.Flush()
}

func health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
