package light

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/httphelper"
)

type Handler interface {
	HandleGet(w http.ResponseWriter, r *http.Request)
	HandleSet(w http.ResponseWriter, r *http.Request)
	DiscoverLightState() (*LightState, error)
}

type handler struct {
	logger       lager.Logger
	httpHelper   httphelper.HttpHelper
	gpio         gpio.Gpio
	gpioLightPin uint
}

func NewHandler(
	logger lager.Logger,
	httpHelper httphelper.HttpHelper,
	gpio gpio.Gpio,
	gpioLightPin uint,
) Handler {

	return &handler{
		httpHelper:   httpHelper,
		logger:       logger,
		gpio:         gpio,
		gpioLightPin: gpioLightPin,
	}
}

type LightState struct {
	StateKnown bool
	LightOn    bool
	ErrorMsg   string
}

func (l LightState) StateString() string {
	if !l.StateKnown {
		return "unknown"
	}
	if l.LightOn {
		return "on"
	} else {
		return "off"
	}
}

func (h handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("received request", lager.Data{"method": r.Method, "url": r.URL})

	ls, err := h.DiscoverLightState()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	b, _ := json.Marshal(ls)
	w.Write(b)
}

func (h handler) DiscoverLightState() (*LightState, error) {
	h.logger.Info("Reading light state")
	state, err := h.gpio.Read(h.gpioLightPin)
	if err != nil {
		h.logger.Error("Error reading light state", err)
		return &LightState{StateKnown: false, LightOn: false}, err
	}
	state = strings.TrimSpace(state)

	lightOn, err := strconv.ParseBool(state)
	if err != nil {
		h.logger.Error("Error parsing light state: %v", err)
		return &LightState{StateKnown: false, LightOn: false}, err
	}

	ls := &LightState{
		StateKnown: true,
		LightOn:    lightOn,
	}
	h.logger.Debug("Light state discovered", lager.Data{"state": ls.StateString()})
	return ls, nil
}

func (h handler) HandleSet(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("received request", lager.Data{"method": r.Method, "url": r.URL})

	err := r.ParseForm()
	if err != nil {
		h.logger.Error("Error parsing form - assuming light should be turned on.", err)

		ls := h.turnLightOn()
		renderLightState(ls, w)

		return
	}

	state := r.Form.Get("state")

	if state == "" {
		h.logger.Info("No state provided - assuming light should be turned on.")
		ls := h.turnLightOn()
		renderLightState(ls, w)
		return
	}

	switch state {
	case "off":
		ls := h.turnLightOff()
		renderLightState(ls, w)
		return
	case "on":
		ls := h.turnLightOn()
		renderLightState(ls, w)
		return
	default:
		h.logger.Info("Invalid state provided - assuming light should be turned on.", lager.Data{"state": state})
		ls := h.turnLightOn()
		renderLightState(ls, w)
		return
	}
}

func renderLightState(ls LightState, w http.ResponseWriter) {
	b, _ := json.Marshal(ls)
	w.Write(b)
}

func (h handler) turnLightOn() LightState {
	h.logger.Info("Turning light on")
	err := h.gpio.WriteHigh(h.gpioLightPin)

	if err != nil {
		h.logger.Error("Error turning light on", err)
		return LightState{
			StateKnown: false,
			LightOn:    false,
			ErrorMsg:   err.Error(),
		}
	}

	h.logger.Info("Light is turned on")
	return LightState{
		StateKnown: true,
		LightOn:    true,
	}
}

func (h handler) turnLightOff() LightState {
	h.logger.Info("Turning light off")
	err := h.gpio.WriteLow(h.gpioLightPin)

	if err != nil {
		h.logger.Error("Error turning light off", err)
		return LightState{
			StateKnown: false,
			LightOn:    false,
			ErrorMsg:   err.Error(),
		}
	}

	h.logger.Info("Light is turned off")
	return LightState{
		StateKnown: true,
		LightOn:    false,
	}
}
