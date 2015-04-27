package garagepi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

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

func (e Executor) handleLightGet(w http.ResponseWriter, r *http.Request) {
	ls, err := e.discoverLightState()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	b, _ := json.Marshal(ls)
	w.Write(b)
}

func (e Executor) discoverLightState() (*LightState, error) {
	e.logger.Log("Reading light state")
	state, err := e.gpio.Read(e.gpioLightPin)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error reading light state: %v", err))
		return &LightState{StateKnown: false, LightOn: false}, err
	}
	state = strings.TrimSpace(state)

	lightOn, err := strconv.ParseBool(state)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error parsing light state: %v", err))
		return &LightState{StateKnown: false, LightOn: false}, err
	}

	ls := &LightState{
		StateKnown: true,
		LightOn:    lightOn,
	}
	e.logger.Log(fmt.Sprintf("Light state: %s", ls.StateString()))
	return ls, nil
}

func (e Executor) handleLightSet(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		e.logger.Log("Error parsing form - assuming light should be turned on.")

		ls := e.turnLightOn()
		renderLightState(ls, w)

		return
	}

	state := r.Form.Get("state")

	if state == "" {
		e.logger.Log("No state provided - assuming light should be turned on.")
		ls := e.turnLightOn()
		renderLightState(ls, w)
		return
	}

	switch state {
	case "off":
		ls := e.turnLightOff()
		renderLightState(ls, w)
		return
	case "on":
		ls := e.turnLightOn()
		renderLightState(ls, w)
		return
	default:
		e.logger.Log(fmt.Sprintf("Invalid state provided (%s) - assuming light should be turned on.", state))
		ls := e.turnLightOn()
		renderLightState(ls, w)
		return
	}
}

func renderLightState(ls LightState, w http.ResponseWriter) {
	b, _ := json.Marshal(ls)
	w.Write(b)
}

func (e Executor) turnLightOn() LightState {
	e.logger.Log(fmt.Sprintf("Turning light on"))
	err := e.gpio.WriteHigh(e.gpioLightPin)

	if err != nil {
		e.logger.Log(fmt.Sprintf("Error turning light on: %v", err))
		return LightState{
			StateKnown: false,
			LightOn:    false,
			ErrorMsg:   err.Error(),
		}
	}

	e.logger.Log(fmt.Sprintf("Light is turned on"))
	return LightState{
		StateKnown: true,
		LightOn:    true,
	}
}

func (e Executor) turnLightOff() LightState {
	e.logger.Log(fmt.Sprintf("Turning light off"))
	err := e.gpio.WriteLow(e.gpioLightPin)

	if err != nil {
		e.logger.Log(fmt.Sprintf("Error turning light off: %v", err))
		return LightState{
			StateKnown: false,
			LightOn:    false,
			ErrorMsg:   err.Error(),
		}
	}

	e.logger.Log(fmt.Sprintf("Light is turned off"))
	return LightState{
		StateKnown: true,
		LightOn:    false,
	}
}
