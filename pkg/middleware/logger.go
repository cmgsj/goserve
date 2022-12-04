package middleware

import (
	"log"
	"net/http"
	"time"
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

func (s *StatusRecorder) Milis() int64 {
	return time.Since(s.Start).Nanoseconds() / 1000 / 1000
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Status:         http.StatusText(http.StatusOK),
		Start:          time.Now(),
	}
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := NewStatusRecorder(w)
		next.ServeHTTP(rec, r)
		if bytesCopied := r.Header.Get("bytes-copied"); bytesCopied != "" {
			log.Printf("%s %s [%s] --> %s [%s] %dms\n", r.Method, r.URL.Path, r.RemoteAddr, rec.Status, bytesCopied, rec.Milis())
		} else {
			log.Printf("%s %s [%s] --> %s %dms\n", r.Method, r.URL.Path, r.RemoteAddr, rec.Status, rec.Milis())
		}
	})
}
