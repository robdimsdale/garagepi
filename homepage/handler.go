package homepage

import (
	"net/http"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/fshelper"
	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/light"
)

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger       lager.Logger
	httpHelper   httphelper.HttpHelper
	fsHelper     fshelper.FsHelper
	lightHandler light.Handler
}

func NewHandler(
	logger lager.Logger,
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
	h.logger.Debug("received request", lager.Data{"method": r.Method, "url": r.URL})

	t, err := h.fsHelper.GetHomepageTemplate()
	if err != nil {
		h.logger.Error("error reading homepage template", err)
		panic(err)
	}

	ls, err := h.lightHandler.DiscoverLightState()
	if err != nil {
		h.logger.Error("error reading light state - rendering homepage without light controls", err)
	}

	t.Execute(w, ls)
}
