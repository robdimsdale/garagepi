package garagepi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	SleepTime        = 500 * time.Millisecond
	GpioReadCommand  = "read"
	GpioWriteCommand = "write"
	GpioLowState     = "0"
	GpioHighState    = "1"
)

type ExecutorConfig struct {
	WebcamHost     string
	WebcamPort     uint
	GpioDoorPin    uint
	GpioLightPin   uint
	GpioExecutable string
}

type Executor struct {
	logger         Logger
	osHelper       OsHelper
	fsHelper       FsHelper
	httpHelper     HttpHelper
	webcamUrl      string
	gpioDoorPin    uint
	gpioLightPin   uint
	gpioExecutable string
}

func NewExecutor(
	logger Logger,
	osHelper OsHelper,
	fsHelper FsHelper,
	httpHelper HttpHelper,
	config ExecutorConfig) *Executor {

	webcamUrl := fmt.Sprintf("http://%s:%d/?action=snapshot&n=", config.WebcamHost, config.WebcamPort)

	return &Executor{
		httpHelper:     httpHelper,
		logger:         logger,
		webcamUrl:      webcamUrl,
		osHelper:       osHelper,
		fsHelper:       fsHelper,
		gpioDoorPin:    config.GpioDoorPin,
		gpioLightPin:   config.GpioLightPin,
		gpioExecutable: config.GpioExecutable,
	}
}

func tostr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}

func (e Executor) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	e.logger.Log("homepage")
	bytes, err := e.fsHelper.GetHomepageTemplateContents()
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error reading homepage template: %v", err))
		panic(err)
	}
	w.Write(bytes)
}

func (e Executor) WebcamHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := e.httpHelper.Get(e.webcamUrl + r.Form.Get("n"))
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error getting image: %v", err))
		if resp == nil {
			e.logger.Log("No image to return")
			return
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error closing image request: %v", err))
		return
	}
	w.Write(body)
}

func (e Executor) ToggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	args := []string{GpioWriteCommand, tostr(e.gpioDoorPin), GpioHighState}
	e.logger.Log("Toggling door")
	_, err := e.executeCommand(e.gpioExecutable, args...)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error executing: '%s %s' - door not toggled (skipping sleep and further executions)", e.gpioExecutable, strings.Join(args, " ")))
		w.Write([]byte("error - door not toggled"))
		return
	} else {
		e.osHelper.Sleep(SleepTime)
		e.executeCommand(e.gpioExecutable, GpioWriteCommand, tostr(e.gpioDoorPin), GpioLowState)
		e.logger.Log("door toggled")
		w.Write([]byte("door toggled"))
		return
	}
}

func (e Executor) GetLightHandler(w http.ResponseWriter, r *http.Request) {
	args := []string{GpioReadCommand, tostr(e.gpioLightPin)}

	e.logger.Log("Reading light state")
	discovered, err := e.executeCommand(e.gpioExecutable, args...)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error executing: '%s %s' - light state unknown", e.gpioExecutable, strings.Join(args, " ")))
		w.Write([]byte("error - light state: unknown"))
	} else {
		state, err := stateNumberToOnOffString(discovered)
		if err != nil {
			w.Write([]byte("error - light state: unknown"))
		} else {
			w.Write([]byte(fmt.Sprintf("light state: %s", state)))
		}
	}
}

func stateNumberToOnOffString(number string) (string, error) {
	switch number {
	case GpioLowState:
		return "off", nil
	case GpioHighState:
		return "on", nil
	default:
		return "", errors.New(fmt.Sprintf("Unrecognized state: %s", number))
	}
}

func onOffStringToStateNumber(onOff string) (string, error) {
	switch onOff {
	case "on":
		return GpioHighState, nil
	case "off":
		return GpioLowState, nil
	default:
		return "", errors.New(fmt.Sprintf("Unrecognized state: %s", onOff))
	}
}

func (e Executor) SetLightHandler(w http.ResponseWriter, r *http.Request) {
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

	e.logger.Log(fmt.Sprintf("Turning light %s", state))
	_, err := e.executeCommand(e.gpioExecutable, args...)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error executing: '%s %s'", e.gpioExecutable, strings.Join(args, " ")))
		w.Write([]byte("error - light state unchanged"))
	} else {
		w.Write([]byte(fmt.Sprintf("light %s", state)))
	}
}

func (e Executor) turnLightOn(w http.ResponseWriter) {
	e.setLightState(w, true)
}

func (e Executor) turnLightOff(w http.ResponseWriter) {
	e.setLightState(w, false)
}

func (e Executor) executeCommand(executable string, arg ...string) (string, error) {
	e.logger.Log(fmt.Sprintf("executing: '%s %s'", executable, strings.Join(arg, " ")))
	out, err := e.osHelper.Exec(executable, arg...)
	if err != nil {
		e.logger.Log(err.Error())
	}
	return out, err
}
