package garagepi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	SleepTime        = 500 * time.Millisecond
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

func (e *Executor) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	e.logger.Log("homepage")
	bytes, err := e.fsHelper.GetHomepageTemplateContents()
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error reading homepage template: %v", err))
		panic(err)
	}
	w.Write(bytes)
}

func (e *Executor) WebcamHandler(w http.ResponseWriter, r *http.Request) {
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

func (e *Executor) ToggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	args := []string{GpioWriteCommand, tostr(e.gpioDoorPin), GpioHighState}
	_, err := e.executeCommand(e.gpioExecutable, args...)
	if err != nil {
		e.logger.Log(fmt.Sprintf("Error executing: '%s %s' - skipping sleep and further executions", e.gpioExecutable, strings.Join(args, " ")))
		w.Write([]byte("error - door not toggled"))
		return
	} else {
		e.osHelper.Sleep(SleepTime)
		e.executeCommand(e.gpioExecutable, GpioWriteCommand, tostr(e.gpioDoorPin), GpioLowState)
		w.Write([]byte("door toggled"))
		return
	}
}

func (e *Executor) LightHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		e.logger.Log("Error parsing form - assuming light should be turned on.")
		e.turnLightOn(w)
		return
	}

	state := r.Form.Get("state")
	switch state {
	case "":
		e.logger.Log("No state provided - assuming light should be turned on.")
		e.turnLightOn(w)
		return
	case "off":
		e.turnLightOff(w)
		return
	case "on":
		e.turnLightOn(w)
		return
	default:
		e.logger.Log(fmt.Sprintf("Invalid state provided (%s) - assuming light should be turned on.", state))
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
		e.handleLightError(w, e.gpioExecutable, args...)
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

func (e Executor) handleLightError(w http.ResponseWriter, executable string, arg ...string) {
	e.logger.Log(fmt.Sprintf("Error executing: '%s %s'", executable, strings.Join(arg, " ")))
	w.Write([]byte("error - light state unchanged"))
}

func (e *Executor) executeCommand(executable string, arg ...string) (string, error) {
	e.logger.Log(fmt.Sprintf("executing: '%s %s'", executable, strings.Join(arg, " ")))
	out, err := e.osHelper.Exec(executable, arg...)
	if err != nil {
		e.logger.Log(err.Error())
	}
	return out, err
}
