package middleware

import (
	"crypto/subtle"
	"net/http"
)

type BasicAuth struct {
	Username, Password string
}

func NewBasicAuth(username, password string) Middleware {
	return BasicAuth{
		Username: username,
		Password: password,
	}
}

func (b BasicAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if !ok {
			setResponseHeader(rw)
			http.Error(rw, "Not Authorized", http.StatusUnauthorized)
		} else {
			if secureCompare(username, b.Username) &&
				secureCompare(password, b.Password) {
				next.ServeHTTP(rw, req)
			} else {
				setResponseHeader(rw)
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
