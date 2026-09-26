package httpresponse

import "net/http"

type (
	responseData struct {
		statusCode int
		bodySize   int
	}
	LoggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

var (
	StatusCodeUninitialized = -1
)

func New(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		responseData: &responseData{
			statusCode: StatusCodeUninitialized,
		},
	}
}

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.bodySize += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

func (r *LoggingResponseWriter) GetStatusCode() int {
	if r.responseData.statusCode == StatusCodeUninitialized {
		panic("no status code set")
	}
	return r.responseData.statusCode
}

func (r *LoggingResponseWriter) GetBodySize() int {
	return r.responseData.bodySize
}
