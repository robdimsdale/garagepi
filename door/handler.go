package door

import (
	"net/http"
	"time"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/oshelper"
)

//go:generate counterfeiter . Handler

var (
	SleepTime = 500 * time.Millisecond
)

type Handler interface {
	HandleToggle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger      lager.Logger
	httpHelper  httphelper.HttpHelper
	osHelper    oshelper.OsHelper
	gpio        gpio.Gpio
	gpioDoorPin uint
}

func NewHandler(
	logger lager.Logger,
	httpHelper httphelper.HttpHelper,
	osHelper oshelper.OsHelper,
	gpio gpio.Gpio,
	gpioDoorPin uint,
) Handler {

	return &handler{
		httpHelper:  httpHelper,
		logger:      logger,
		gpio:        gpio,
		gpioDoorPin: gpioDoorPin,
		osHelper:    osHelper,
	}
}

func (h handler) HandleToggle(w http.ResponseWriter, r *http.Request) {
	err := h.gpio.WriteHigh(h.gpioDoorPin)
	if err != nil {
		h.logger.Error("error toggling door. Skipping sleep and further executions", err)
		w.Write([]byte("error - door not toggled"))
		return
	} else {
		h.osHelper.Sleep(SleepTime)

		err := h.gpio.WriteLow(h.gpioDoorPin)
		if err != nil {
			h.logger.Error("error toggling door", err)
		}

		h.logger.Info("door toggled")
		w.Write([]byte("door toggled"))
		return
	}
}
