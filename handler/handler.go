package handler

import (
	"net/http"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/middleware"
)

func newHandler(
	mux http.Handler,
	logger lager.Logger,
	username string,
	password string,
) http.Handler {
	if username == "" && password == "" {
		return middleware.Chain{
			middleware.NewPanicRecovery(logger),
			middleware.NewLogger(logger),
		}.Wrap(mux)
	} else {
		return middleware.Chain{
			middleware.NewPanicRecovery(logger),
			middleware.NewLogger(logger),
			middleware.NewBasicAuth("username", "password"),
		}.Wrap(mux)
	}
}
