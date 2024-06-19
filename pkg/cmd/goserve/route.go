package goserve

import (
	"fmt"
	"net/http"
)

type route struct {
	patterns    []string
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

		for _, pattern := range route.patterns {
			mux.Handle(pattern, route.handler)

			configs = append(configs, config{
				key: fmt.Sprintf("%s\t->\t%s", pattern, route.description),
			})
		}
	}

	return printConfigs(configs)
}

func health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
