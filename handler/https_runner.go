package handler

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type httpsRunner struct {
	port      uint
	logger    lager.Logger
	handler   http.Handler
	tlsConfig *tls.Config
}

func NewHTTPSRunner(
	port uint,
	logger lager.Logger,
	handler http.Handler,
	keyFile string,
	certFile string,
	caFile string,
	username string,
	password string,
	cookieHandler *securecookie.SecureCookie,
) ifrit.Runner {

	tlsConfig := createTLSConfig(keyFile, certFile, caFile)

	var h http.Handler
	if username == "" && password == "" {
		h = newHandler(handler, logger)
	} else {
		h = newSessionAuthHandler(handler, logger, username, password, cookieHandler)
	}

	return &httpsRunner{
		port:      port,
		logger:    logger,
		handler:   h,
		tlsConfig: tlsConfig,
	}
}

func (r httpsRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var listener net.Listener
	var err error

	listener, err = tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.port), r.tlsConfig)
	if err != nil {
		return err
	}
	r.logger.Info("HTTPS server listening on port", lager.Data{"port": r.port})

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

func createTLSConfig(keyFile string, certFile string, caFile string) *tls.Config {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if caFile != "" {
		// Load CA cert
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.BuildNameToCertificate()
	return tlsConfig
}
