package garagepi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type LightState struct {
	StateKnown bool
	LightOn    bool
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

	w.Write([]byte(fmt.Sprintf("light state: %s", ls.StateString())))
}

func (e Executor) discoverLightState() (*LightState, error) {
	args := []string{GpioReadCommand, tostr(e.gpioLightPin)}
	e.logger.Log("Reading light state")
	state, err := e.executeCommand(e.gpioExecutable, args...)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error executing: '%s %s' - light state unknown", e.gpioExecutable, strings.Join(args, " ")))
		return &LightState{StateKnown: false}, err
	}
	state = strings.TrimSpace(state)

	lightOn, err := strconv.ParseBool(state)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error parsing light state: %v", err))
		return &LightState{StateKnown: false}, err
	}

	ls := &LightState{
		StateKnown: true,
		LightOn:    lightOn,
	}
	e.logger.Log(fmt.Sprintf("Light state: %s", ls.StateString()))
	return ls, nil
}

func (e Executor) handleLightState(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		e.logger.Log("Error parsing form - assuming light should be turned on.")
		e.turnLightOn(w)
		return
	}

	state := r.Form.Get("state")

	if state == "" {
		e.logger.Log("No state provided - assuming light should be turned on.")
		e.turnLightOn(w)
		return
	}

	gpioState, err := onOffStringToStateNumber(state)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Invalid state provided (%s) - assuming light should be turned on.", state))
		e.turnLightOn(w)
		return
	}

	switch gpioState {
	case GpioLowState:
		e.turnLightOff(w)
		return
	case GpioHighState:
		e.turnLightOn(w)
		return
	}
}

func (e Executor) turnLightOn(w http.ResponseWriter) {
	e.setLightState(w, true)
}

func (e Executor) turnLightOff(w http.ResponseWriter) {
	e.setLightState(w, false)
}

func (e Executor) setLightState(w http.ResponseWriter, stateOn bool) {
	var state string
	var args []string
	if stateOn {
		state = "on"
		args = []string{GpioWriteCommand, tostr(e.gpioLightPin), GpioHighState}
	} else {
		state = "off"
		args = []string{GpioWriteCommand, tostr(e.gpioLightPin), GpioLowState}
	}

	e.logger.Log(fmt.Sprintf("Setting light state to %s", state))
	_, err := e.executeCommand(e.gpioExecutable, args...)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error executing: '%s %s'", e.gpioExecutable, strings.Join(args, " ")))
		w.Write([]byte("error - light state unchanged"))
	} else {
		e.logger.Log(fmt.Sprintf("Light state: %s", state))
		w.Write([]byte(fmt.Sprintf("light state: %s", state)))
	}
}