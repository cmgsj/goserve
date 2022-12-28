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

type StatusRecorder struct {
	http.ResponseWriter
	StatusCode int
	Status     string
	Start      time.Time
}

func (s *StatusRecorder) WriteHeader(code int) {
	s.StatusCode = code
	s.Status = http.StatusText(code)
	s.ResponseWriter.WriteHeader(code)
}

func (s *StatusRecorder) Duration() time.Duration {
	return time.Since(s.Start)
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Status:         http.StatusText(http.StatusOK),
		Start:          time.Now(),
	}
}

func Logger(out io.Writer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := NewStatusRecorder(w)
		next.ServeHTTP(rec, r)
		fmt.Fprintf(out, "%s %s %s [%s] --> %s [%s] %dms\n",
			time.Now().Format("2006/01/02 15:04:05"),
			r.Method, r.URL.Path, r.RemoteAddr, rec.Status,
			r.Header.Get(BytesCopiedHeader),
			rec.Duration().Milliseconds())
	})
}
