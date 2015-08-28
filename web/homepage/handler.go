package homepage

import (
	"html/template"
	"net/http"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/api/light"
	"github.com/robdimsdale/garagepi/web/login"
)

//go:generate counterfeiter . Handler

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger       lager.Logger
	templates    *template.Template
	lightHandler light.Handler
	loginHandler login.Handler
}

func NewHandler(
	logger lager.Logger,
	templates *template.Template,
	lightHandler light.Handler,
	loginHandler login.Handler,
) Handler {
	return &handler{
		logger:       logger,
		templates:    templates,
		lightHandler: lightHandler,
		loginHandler: loginHandler,
	}
}

func (h handler) Handle(w http.ResponseWriter, r *http.Request) {
	ls, err := h.lightHandler.DiscoverLightState()
	if err != nil {
		h.logger.Error("error reading light state - rendering homepage without light controls", err)
	}

	h.templates.ExecuteTemplate(w, "homepage", ls)
}
