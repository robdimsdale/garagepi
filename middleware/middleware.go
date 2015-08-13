package middleware

import "net/http"

type Middleware interface {
	Wrap(http.Handler) http.Handler
}
