package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/pivotal-golang/lager"
)

type BasicAuth struct {
	Username, Password string
	logger             lager.Logger
}

func NewBasicAuth(username, password string, logger lager.Logger) Middleware {
	return BasicAuth{
		Username: username,
		Password: password,
		logger:   logger,
	}
}

func (b BasicAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			b.logger.Debug("no authentication provided")
			setResponseHeader(rw)
			http.Error(rw, "Not Authorized", http.StatusUnauthorized)
		} else {
			if secureCompare(username, b.Username) &&
				secureCompare(password, b.Password) {
				b.logger.Debug("successful authorization", lager.Data{"username": username})
				next.ServeHTTP(rw, req)
			} else {
				setResponseHeader(rw)
				b.logger.Debug("unsuccessful authorization", lager.Data{"username": username})
				http.Error(rw, "Not Authorized", http.StatusForbidden)
			}
		}
	})
}

func secureCompare(a, b string) bool {
	x := []byte(a)
	y := []byte(b)
	return subtle.ConstantTimeCompare(x, y) == 1
}

func setResponseHeader(rw http.ResponseWriter) {
	rw.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
}
