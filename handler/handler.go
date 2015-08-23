package handler

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/middleware"
)

func newSessionAuthHandler(
	mux http.Handler,
	logger lager.Logger,
	username string,
	password string,
	cookieHandler *securecookie.SecureCookie,
) http.Handler {
	return middleware.Chain{
		middleware.NewPanicRecovery(logger),
		middleware.NewLogger(logger),
		middleware.NewSessionAuth(username, password, logger, cookieHandler),
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
