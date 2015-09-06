package middleware

import (
	"net/http"

	"github.com/pivotal-golang/lager"
)

type panicRecovery struct {
	logger lager.Logger
}

func NewPanicRecovery(logger lager.Logger) Middleware {
	return &panicRecovery{logger}
}

func (p panicRecovery) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				p.logger.Error("Panic while serving request", nil, lager.Data{
					"request":   loggedRequest(*req),
					"panicInfo": panicInfo,
				})
			}
		}()
		next.ServeHTTP(rw, req)
	})
}
