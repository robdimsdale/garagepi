package door

import (
	"fmt"
	"net/http"
	"time"

	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/oshelper"
)

var (
	SleepTime = 500 * time.Millisecond
)

type Handler interface {
	HandleToggle(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	logger      logger.Logger
	httpHelper  httphelper.HttpHelper
	osHelper    oshelper.OsHelper
	gpio        gpio.Gpio
	gpioDoorPin uint
}

func NewHandler(
	logger logger.Logger,
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
	h.logger.Log(fmt.Sprintf("%s request to %v", r.Method, r.URL))

	err := h.gpio.WriteHigh(h.gpioDoorPin)
	if err != nil {
		h.logger.Log(fmt.Sprintf("Error toggling door (skipping sleep and further executions): %v", err))
		w.Write([]byte("error - door not toggled"))
		return
	} else {
		h.osHelper.Sleep(SleepTime)

		err := h.gpio.WriteLow(h.gpioDoorPin)
		if err != nil {
			h.logger.Log(fmt.Sprintf("Error toggling door: %v", err))
		}

		h.logger.Log("door toggled")
		w.Write([]byte("door toggled"))
		return
	}
}
