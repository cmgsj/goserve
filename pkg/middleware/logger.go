package middleware

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type httpLogger struct {
	base http.Handler
	out  io.Writer
}

func NewHTTPLogger(base http.Handler, out io.Writer) http.Handler {
	return &httpLogger{base: base, out: out}
}

func (l *httpLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := newHTTPRecorder(w)
	start := time.Now()
	l.base.ServeHTTP(rec, r)
	delta := time.Since(start)
	fmt.Fprintf(l.out, "%s %s %s %s -> %s [%s] %s\n",
		start.Format("2006/01/02 15:04:05"), r.Method, r.URL.Path, r.RemoteAddr, rec.status, formatDuration(delta), r.Header.Get("X-Bytes-Copied"))
}

type httpRecorder struct {
	http.ResponseWriter
	code   int
	status string
}

func newHTTPRecorder(w http.ResponseWriter) *httpRecorder {
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

func formatDuration(d time.Duration) string {
	var (
		unit   string
		factor int64
		n      = d.Nanoseconds()
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
