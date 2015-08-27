package webcam

import (
	"io/ioutil"
	"net/http"

	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . Handler

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger    lager.Logger
	webcamURL string
}

func NewHandler(
	logger lager.Logger,
	webcamURL string,
) Handler {

	return &handler{
		logger:    logger,
		webcamURL: webcamURL,
	}
}

func (h handler) Handle(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(h.webcamURL + r.Form.Get("n"))
	if err != nil {
		h.logger.Error("error getting image", err)
		if resp == nil {
			h.logger.Info("no image to return")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.Info("Bad upstream status code", lager.Data{"statusCode": resp.StatusCode})
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("error reading returned image", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.Write(body)
}
