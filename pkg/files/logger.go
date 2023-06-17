package files

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

func RequestLogger(base http.Handler, out io.Writer) http.Handler {
	return &httpLogger{base: base, out: out}
}

func (l *httpLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := &httpRecorder{ResponseWriter: w}
	start := time.Now()
	l.base.ServeHTTP(rec, r)
	delta := time.Since(start)
	fmt.Fprintf(l.out, "%s %s %s %s -> %s [%s] %s\n",
		start.Format("2006/01/02 15:04:05"), r.Method, r.URL.Path, r.RemoteAddr, rec.status, formatDuration(delta), r.Header.Get(BytesCopiedHeader))
}

type httpRecorder struct {
	http.ResponseWriter
	code   int
	status string
}

func (r *httpRecorder) WriteHeader(code int) {
	r.code = code
	r.status = http.StatusText(code)
	r.ResponseWriter.WriteHeader(code)
}
