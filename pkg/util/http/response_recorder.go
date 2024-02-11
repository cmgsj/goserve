package http

import "net/http"

type ResponseRecorder interface {
	http.ResponseWriter
	Status() string
	StatusCode() int
}

func NewResponseRecorder(w http.ResponseWriter) ResponseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.statusCode = code
}

func (r *responseRecorder) Status() string {
	return http.StatusText(r.statusCode)
}

func (r *responseRecorder) StatusCode() int {
	return r.statusCode
}
