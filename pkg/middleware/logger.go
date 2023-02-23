package middleware

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	StatusCode int
	Status     string
	StartTime  time.Time
	TimeDelta  time.Duration
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Status:         http.StatusText(http.StatusOK),
	}
}

func (s *statusRecorder) WriteHeader(code int) {
	s.ResponseWriter.WriteHeader(code)
	s.StatusCode = code
	s.Status = http.StatusText(code)
}

func (s *statusRecorder) Start() { s.StartTime = time.Now() }

func (s *statusRecorder) Stop() { s.TimeDelta = time.Since(s.StartTime) }

func Logger(next http.Handler, outWriter io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := newStatusRecorder(w)
		rec.Start()
		next.ServeHTTP(rec, r)
		rec.Stop()
		fmt.Fprintf(outWriter, "%s %s %s %s -> %s [%s] %s\n",
			rec.StartTime.Format("2006/01/02 15:04:05"), r.Method, r.URL.Path, r.RemoteAddr, rec.Status,
			formatDuration(rec.TimeDelta), r.Header.Get("bytes-copied"))
	})
}

func formatDuration(t time.Duration) string {
	var (
		unit   string
		factor int64
		n      = t.Nanoseconds()
	)
	if factor = 60 * 1000 * 1000 * 1000; n >= factor {
		unit = "min"
	} else if factor = 1000 * 1000 * 1000; n >= factor {
		unit = "s"
	} else if factor = 1000 * 1000; n >= factor {
		unit = "ms"
	} else if factor = 1000; n >= factor {
		unit = "µs"
	} else {
		unit = "ns"
		factor = 1
	}
	return fmt.Sprintf("%.2f%s", float64(n)/float64(factor), unit)
}
