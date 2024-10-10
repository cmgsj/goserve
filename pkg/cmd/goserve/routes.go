package goserve

import (
	"fmt"
	"net/http"
	"strings"
)

type route struct {
	pattern     string
	description string
	handler     http.Handler
	disabled    bool
}

func registerRoutes(mux *http.ServeMux, routes []route) error {
	var configs []config

	for _, route := range routes {
		if route.disabled {
			continue
		}

		mux.Handle(route.pattern, route.handler)

		configs = append(configs, config{
			key: fmt.Sprintf("%s\t->\t%s", route.pattern, route.description),
		})
	}

	return printConfigs(configs)
}

func redirect(url string, code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dst := url

		if r.URL.RawQuery != "" {
			sep := "?"
			if strings.Contains(dst, "?") {
				sep = "&"
			}
			dst += sep + r.URL.RawQuery
		}

		http.Redirect(w, r, dst, code)
	})
}

func health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
