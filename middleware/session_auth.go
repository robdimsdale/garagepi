package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
)

type sessionAuth struct {
	username, password string
	logger             lager.Logger
	cookieHandler      *securecookie.SecureCookie
}

func NewSessionAuth(
	username string,
	password string,
	logger lager.Logger,
	cookieHandler *securecookie.SecureCookie,
) Middleware {
	return sessionAuth{
		username:      username,
		password:      password,
		logger:        logger,
		cookieHandler: cookieHandler,
	}
}

func (s sessionAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if unauthenticatedAccessAllowedForURL(req.URL.Path) {
			s.logger.Debug(
				"unauthenticated access allowed for URL",
				lager.Data{"url": req.URL.Path},
			)
			next.ServeHTTP(rw, req)
		} else if s.isLoggedIn(req) {
			s.logger.Debug("successful authorization via session")
			next.ServeHTTP(rw, req)
		} else {
			s.logger.Debug("not logged in - redirecting")
			http.Redirect(rw, req, "/login", http.StatusFound)
		}
	})
}

func unauthenticatedAccessAllowedForURL(url string) bool {
	openURLs := []string{"/login", "/static"}

	for _, u := range openURLs {
		if strings.HasPrefix(url, u) {
			return true
		}
	}
	return false
}

func (s sessionAuth) isLoggedIn(request *http.Request) bool {
	var username, password string
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = s.cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			username = cookieValue["name"]
			password = cookieValue["password"]
		}
	}
	return secureCompare(username, s.username) && secureCompare(password, s.password)
}

func secureCompare(a, b string) bool {
	x := []byte(a)
	y := []byte(b)
	return subtle.ConstantTimeCompare(x, y) == 1
}
