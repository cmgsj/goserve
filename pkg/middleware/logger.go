package middleware

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BytesCopiedHeader = "bytes-copied"
)

type statusRecorder struct {
	http.ResponseWriter
	StatusCode int
	Status     string
}

func (s *statusRecorder) WriteHeader(code int) {
	s.StatusCode = code
	s.Status = http.StatusText(code)
	s.ResponseWriter.WriteHeader(code)
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Status:         http.StatusText(http.StatusOK),
	}
}

func Logger(out io.Writer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := newStatusRecorder(w)
		start := time.Now()
		next.ServeHTTP(rec, r)
		end := time.Now()
		fmt.Fprintf(out, "%s %s %s [%s] --> %s [%s] %dms\n",
			end.Format("2006/01/02 15:04:05"),
			r.Method, r.URL.Path, r.RemoteAddr, rec.Status,
			r.Header.Get(BytesCopiedHeader),
			end.Sub(start).Milliseconds(),
		)
	})
}
