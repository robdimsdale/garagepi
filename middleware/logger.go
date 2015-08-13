package middleware

import (
	"net/http"
	"strconv"

	"github.com/pivotal-golang/lager"
)

type Logger struct {
	logger lager.Logger
}

func NewLogger(logger lager.Logger) Middleware {
	return Logger{
		logger: logger,
	}
}

func (l Logger) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		loggingResponseWriter := responseWriter{
			rw,
			[]byte{},
			0,
		}
		next.ServeHTTP(&loggingResponseWriter, req)

		requestCopy := *req
		requestCopy.Header["Authorization"] = nil

		response := map[string]interface{}{
			"Header":     loggingResponseWriter.Header(),
			"Body":       string(loggingResponseWriter.body),
			"StatusCode": loggingResponseWriter.statusCode,
		}

		l.logger.Debug("", lager.Data{
			"request":  requestCopy,
			"response": response,
		})
	})
}

type responseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.Header().Set("Content-Length", strconv.Itoa(len(b)))

	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	size, err := rw.ResponseWriter.Write(b)
	rw.body = b
	return size, err
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.statusCode = s
	rw.ResponseWriter.WriteHeader(s)
}
