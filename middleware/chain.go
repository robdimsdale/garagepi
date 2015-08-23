package middleware

import "net/http"

type Chain []Middleware

func (mc Chain) Wrap(next http.Handler) http.Handler {
	var chain = next

	middlewareCount := len(mc)
	for i := middlewareCount - 1; i >= 0; i-- { // iterate backwards
		chain = mc[i].Wrap(chain)
	}
	return chain
}
