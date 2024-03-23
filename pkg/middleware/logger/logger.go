package logger

import (
	"log/slog"
	"net/http"
	"time"

	utilhttp "github.com/cmgsj/goserve/pkg/util/http"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := utilhttp.NewResponseRecorder(w)

		start := time.Now()

		defer func() {
			slog.Info(
				r.Method+" "+r.URL.Path,
				"address", r.RemoteAddr,
				"status", http.StatusText(recorder.StatusCode()),
				"duration", time.Since(start),
			)
		}()

		next.ServeHTTP(recorder, r)
	})
}
