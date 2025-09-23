package http

import "net/http"

type ResponseRecorder interface {
	http.ResponseWriter
	StatusCode() int
	BytesWritten() int64
}

func NewResponseRecorder(w http.ResponseWriter) ResponseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

type responseRecorder struct {
	http.ResponseWriter

	statusCode   int
	bytesWritten int64
}

func (r *responseRecorder) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.statusCode = code
}

func (r *responseRecorder) Write(content []byte) (int, error) {
	n, err := r.ResponseWriter.Write(content)
	r.bytesWritten += int64(n)

	return n, err
}

func (r *responseRecorder) StatusCode() int {
	return r.statusCode
}

func (r *responseRecorder) BytesWritten() int64 {
	return r.bytesWritten
}
