package logging

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/cmgsj/goserve/pkg/files"
	middlewarehttp "github.com/cmgsj/goserve/pkg/middleware/http"
)

func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := middlewarehttp.NewResponseRecorder(w)

		start := time.Now()

		defer func() {
			slog.Info(
				r.Method+" "+r.URL.Path,
				"address", r.RemoteAddr,
				"status", http.StatusText(recorder.StatusCode()),
				"size", files.FormatSizeMetricUnits(float64(recorder.BytesWritten()), files.ShortestLength),
				"duration", time.Since(start),
			)
		}()

		next.ServeHTTP(recorder, r)
	})
}
