package middleware

import (
	"net/http"

	"github.com/pivotal-golang/lager"
)

type PanicRecovery struct {
	logger lager.Logger
}

func NewPanicRecovery(logger lager.Logger) Middleware {
	return &PanicRecovery{logger}
}

func (p PanicRecovery) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				p.logger.Error("Panic while serving request", nil, lager.Data{
					"request":   req,
					"panicInfo": panicInfo,
				})
			}
		}()
		next.ServeHTTP(rw, req)
	})
}
