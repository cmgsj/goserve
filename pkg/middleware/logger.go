package middleware

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type httpRecorder struct {
	http.ResponseWriter
	code   int
	status string
}

func newHttpRecorder(w http.ResponseWriter) *httpRecorder {
	return &httpRecorder{
		ResponseWriter: w,
		code:           http.StatusOK,
		status:         http.StatusText(http.StatusOK),
	}
}

func (r *httpRecorder) WriteHeader(code int) {
	r.code = code
	r.status = http.StatusText(code)
	r.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler, out io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := newHttpRecorder(w)
		start := time.Now()
		next.ServeHTTP(rec, r)
		delta := time.Since(start)
		fmt.Fprintf(out, "%s %s %s %s -> %s [%s] %s\n",
			start.Format("2006/01/02 15:04:05"), r.Method, r.URL.Path, r.RemoteAddr, rec.status, formatDuration(delta), r.Header.Get("X-Bytes-Copied"))
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
