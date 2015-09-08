package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
)

type auth struct {
	username, password string
	logger             lager.Logger
	cookieHandler      *securecookie.SecureCookie
}

func NewAuth(
	username string,
	password string,
	logger lager.Logger,
	cookieHandler *securecookie.SecureCookie,
) Middleware {
	return auth{
		username:      username,
		password:      password,
		logger:        logger,
		cookieHandler: cookieHandler,
	}
}

func (s auth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if s.unauthenticatedAccessAllowedForURL(req.URL.Path) ||
			s.validSession(req) ||
			s.validBasicAuth(req) {
			next.ServeHTTP(rw, req)
		} else {
			s.logger.Debug("not logged in - redirecting")
			http.Redirect(rw, req, "/login", http.StatusFound)
		}
	})
}

func (s auth) unauthenticatedAccessAllowedForURL(url string) bool {
	openURLs := []string{"/login", "/static"}

	for _, u := range openURLs {
		if strings.HasPrefix(url, u) {
			s.logger.Debug("unauthenticated access allowed for URL", lager.Data{"url": url})
			return true
		}
	}
	s.logger.Debug("authenticated access required for URL", lager.Data{"url": url})
	return false
}

func (s auth) validBasicAuth(request *http.Request) bool {
	username, password, ok := request.BasicAuth()

	validated := ok &&
		secureCompare(username, s.username) &&
		secureCompare(password, s.password)

	if validated {
		s.logger.Debug("successfully validated via basic auth")
		return true
	}

	s.logger.Debug("failed validation via basic auth")
	return false
}

func (s auth) validSession(request *http.Request) bool {
	var username, password string
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		err = s.cookieHandler.Decode("session", cookie.Value, &cookieValue)
		if err == nil {
			username = cookieValue["name"]
			password = cookieValue["password"]
		}
	}

	validated := secureCompare(username, s.username) &&
		secureCompare(password, s.password)

	if validated {
		s.logger.Debug("successfully validated via session")
		return true
	}
	s.logger.Debug("failed validation via session")
	return false
}

func secureCompare(a, b string) bool {
	x := []byte(a)
	y := []byte(b)
	return subtle.ConstantTimeCompare(x, y) == 1
}
