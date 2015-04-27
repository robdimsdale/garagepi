package homepage

import (
	"fmt"
	"net/http"

	"github.com/robdimsdale/garagepi/fshelper"
	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/light"
	"github.com/robdimsdale/garagepi/logger"
)

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger       logger.Logger
	httpHelper   httphelper.HttpHelper
	fsHelper     fshelper.FsHelper
	lightHandler light.Handler
}

func NewHandler(
	logger logger.Logger,
	httpHelper httphelper.HttpHelper,
	fsHelper fshelper.FsHelper,
	lightHandler light.Handler,
) Handler {

	return &handler{
		httpHelper:   httpHelper,
		logger:       logger,
		fsHelper:     fsHelper,
		lightHandler: lightHandler,
	}
}

func (h handler) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Log(fmt.Sprintf("%s request to %v", r.Method, r.URL))

	t, err := h.fsHelper.GetHomepageTemplate()
	if err != nil {
		h.logger.Log(fmt.Sprintf("Error reading homepage template: %v", err))
		panic(err)
	}

	ls, err := h.lightHandler.DiscoverLightState()
	if err != nil {
		h.logger.Log("Error reading light state - rendering homepage without light controls")
	}

	t.Execute(w, ls)
}
