package goserve

import (
	"fmt"
	"net/http"
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
