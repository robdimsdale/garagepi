package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type httpsEnforcer struct {
	httpsPort uint
}

func NewHTTPSEnforcer(httpsPort uint) Middleware {
	return httpsEnforcer{
		httpsPort: httpsPort,
	}
}

func (h httpsEnforcer) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		redirectTo, err := url.Parse(req.URL.String())
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}

		hostWithoutPort := (strings.Split(req.Host, ":"))[0]
		newHost := fmt.Sprintf("%s:%d", hostWithoutPort, h.httpsPort)

		redirectTo.Host = newHost
		redirectTo.Scheme = "https"

		http.Redirect(rw, req, redirectTo.String(), http.StatusFound)
	})
}
