package handler

import (
	"net/http"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/middleware"
)

func newBasicAuthHandler(
	mux http.Handler,
	logger lager.Logger,
	username string,
	password string,
) http.Handler {
	return middleware.Chain{
		middleware.NewPanicRecovery(logger),
		middleware.NewLogger(logger),
		middleware.NewBasicAuth("username", "password"),
	}.Wrap(mux)
}

func newHandler(
	mux http.Handler,
	logger lager.Logger,
) http.Handler {
	return middleware.Chain{
		middleware.NewPanicRecovery(logger),
		middleware.NewLogger(logger),
	}.Wrap(mux)
}

func newForceHTTPSHandler(
	mux http.Handler,
	logger lager.Logger,
	httpsPort uint,
) http.Handler {
	return middleware.Chain{
		middleware.NewPanicRecovery(logger),
		middleware.NewLogger(logger),
		middleware.NewHTTPSEnforcer(httpsPort),
	}.Wrap(mux)
}
