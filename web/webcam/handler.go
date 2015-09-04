package webcam

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . Handler

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger lager.Logger
	proxy  httputil.ReverseProxy
}

func NewHandler(
	logger lager.Logger,
	webcamHost string,
) Handler {
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = webcamHost
		req.URL.Path = "/"
		req.URL.RawQuery = "action=stream"
	}

	flushInterval, err := time.ParseDuration("10ms")
	if err != nil {
		logger.Fatal("golang broke", err)
	}

	proxy := httputil.ReverseProxy{
		Director:      director,
		FlushInterval: flushInterval,
		ErrorLog:      log.New(ioutil.Discard, "", 0),
	}

	return &handler{
		logger: logger,
		proxy:  proxy,
	}
}

func (h handler) Handle(w http.ResponseWriter, r *http.Request) {
	h.proxy.ServeHTTP(w, r)
}
