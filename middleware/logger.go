package middleware

import (
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
		if l.skipLoggingForUrl(req.URL.Path) {
			next.ServeHTTP(rw, req)
		} else {
			loggingResponseWriter := responseWriter{
				rw,
				[]byte{},
				0,
				0,
			}
			next.ServeHTTP(&loggingResponseWriter, req)

			requestCopy := *req
			requestCopy.Header["Authorization"] = nil

			response := map[string]interface{}{
				"Header":     loggingResponseWriter.Header(),
				"StatusCode": loggingResponseWriter.statusCode,
				"Size":       loggingResponseWriter.size,
			}

			l.logger.Debug("", lager.Data{
				"request":  fromHTTPRequest(requestCopy),
				"response": response,
			})
		}
	})
}

func (l logger) skipLoggingForUrl(url string) bool {
	openURLs := []string{"/webcam"}

	for _, u := range openURLs {
		if strings.HasPrefix(url, u) {
			l.logger.Debug("skiping logging for URL", lager.Data{"url": url})
			return true
		}
	}
	return false
}

type responseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
	size       int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	size, err := rw.ResponseWriter.Write(b)
	rw.body = append(rw.body, b...)
	rw.size += size

	rw.Header().Set("Content-Length", strconv.Itoa(rw.size))

	return size, err
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.statusCode = s
	rw.ResponseWriter.WriteHeader(s)
}

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
	var form, postForm url.Values
	if req.Form != nil {
		form = sanitizeCredentialsFromForm(req.Form)
	}

	if req.PostForm != nil {
		postForm = sanitizeCredentialsFromForm(req.PostForm)
	}

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
		Form:             form,
		PostForm:         postForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
		TLS:              req.TLS,
	}
}

func sanitizeCredentialsFromForm(form url.Values) url.Values {
	form.Set("password", "***")
	return form
}
