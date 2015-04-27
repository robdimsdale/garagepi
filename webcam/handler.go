package webcam

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/logger"
)

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger     logger.Logger
	httpHelper httphelper.HttpHelper
	webcamUrl  string
}

func NewHandler(
	logger logger.Logger,
	httpHelper httphelper.HttpHelper,
	webcamHost string,
	webcamPort uint,
) Handler {

	webcamUrl := fmt.Sprintf("http://%s:%d/?action=snapshot&n=", webcamHost, webcamPort)

	return &handler{
		httpHelper: httpHelper,
		logger:     logger,
		webcamUrl:  webcamUrl,
	}
}

func (h handler) Handle(w http.ResponseWriter, r *http.Request) {
	resp, err := h.httpHelper.Get(h.webcamUrl + r.Form.Get("n"))
	if err != nil {
		h.logger.Log(fmt.Sprintf("Error getting image: %v", err))
		if resp == nil {
			h.logger.Log("No image to return")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.logger.Log(fmt.Sprintf("Error closing image request: %v", err))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.Write(body)
}
