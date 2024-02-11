package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	utilhttp "github.com/cmgsj/goserve/pkg/util/http"
)

func Log(base http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := utilhttp.NewResponseRecorder(w)

		start := time.Now()

		base.ServeHTTP(recorder, r)

		delta := time.Since(start)

		slog.Info(
			fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			"address", r.RemoteAddr,
			"status", recorder.Status(),
			"duration", delta,
		)
	})
}
