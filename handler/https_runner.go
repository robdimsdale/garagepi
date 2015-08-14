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
) ifrit.Runner {
	tlsConfig := createTLSConfig(keyFile, certFile, caFile)
	return &httpsRunner{
		port:      port,
		logger:    logger,
		handler:   handler,
		tlsConfig: tlsConfig,
	}
}

func (r httpsRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var listener net.Listener
	var err error

	listener, err = tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.port), r.tlsConfig)
	if err != nil {
		return err
	} else {
		r.logger.Info("web server listening on port", lager.Data{"port": r.port})
	}

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
