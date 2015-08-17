package middleware

import (
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pivotal-golang/lager"
)

type logger struct {
	logger lager.Logger
}

func NewLogger(l lager.Logger) Middleware {
	return logger{
		logger: l,
	}
}

func (l logger) Wrap(next http.Handler) http.Handler {
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
			"StatusCode": loggingResponseWriter.statusCode,
		}

		l.logger.Debug("", lager.Data{
			"request":  fromHTTPRequest(requestCopy),
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

// Golang 1.5 introduces a http.Request.Cancel field,
// of type <-chan struct{} which the lager library fails to deal with.
// We introduced loggableHTTPRequest as a way of handling this
// until such time as lager can deal with it.
//
// Once lager can handle the request directly, remove this struct.
// #101259402
type LoggableHTTPRequest struct {
	Method           string
	URL              *url.URL
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	Body             io.ReadCloser
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	TLS              *tls.ConnectionState
}

func fromHTTPRequest(req http.Request) LoggableHTTPRequest {
	return LoggableHTTPRequest{
		Method:           req.Method,
		URL:              req.URL,
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           req.Header,
		Body:             req.Body,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Close:            req.Close,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
		TLS:              req.TLS,
	}
}
