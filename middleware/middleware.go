package middleware

import "net/http"

//go:generate counterfeiter . Middleware

type Middleware interface {
	Wrap(http.Handler) http.Handler
}
