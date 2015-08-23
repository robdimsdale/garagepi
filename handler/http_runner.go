package handler

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type httpRunner struct {
	port          uint
	logger        lager.Logger
	handler       http.Handler
	cookieHandler securecookie.SecureCookie
}

func NewHTTPRunner(
	port uint,
	logger lager.Logger,
	handler http.Handler,
	forceHTTPS bool,
	httpsPort uint,
	username string,
	password string,
	cookieHandler *securecookie.SecureCookie,
) ifrit.Runner {

	var h http.Handler
	if forceHTTPS {
		h = newForceHTTPSHandler(handler, logger, httpsPort)
	} else if username != "" && password != "" {
		h = newSessionAuthHandler(handler, logger, username, password, cookieHandler)
	} else {
		h = newHandler(handler, logger)
	}

	return &httpRunner{
		port:    port,
		logger:  logger,
		handler: h,
	}
}

func (r httpRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var listener net.Listener
	var err error

	listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.port))
	if err != nil {
		return err
	}
	r.logger.Info("HTTP server listening on port", lager.Data{"port": r.port})

	errChan := make(chan error)
	go func() {
		err := http.Serve(listener, r.handler)
		if err != nil {
			errChan <- err
		}
	}()

	close(ready)

	select {
	case <-signals:
		return listener.Close()
	case err := <-errChan:
		return err
	}
}
