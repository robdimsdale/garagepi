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
		e.logger.Log("Error reading homepage template: " + err.Error())
		panic(err)
	}
	w.Write(bytes)
}

func (e *Executor) WebcamHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := e.httpHelper.Get(e.webcamUrl + r.Form.Get("n"))
	if err != nil {
		e.logger.Log("Error getting image: " + err.Error())
		if resp == nil {
			return
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.logger.Log("Error closing image request: " + err.Error())
		return
	}
	w.Write(body)
}

func (e *Executor) ToggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	_, err := e.executeCommand(e.gpioExecutable, GpioWriteCommand, tostr(e.gpioDoorPin), GpioHighState)
	if err != nil {
		e.logger.Log("Error occured while executing " + e.gpioExecutable + " - skipping sleep and further executions")
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
	lightOn := true // All errors result in light being turned on
	err := r.ParseForm()
	if err != nil {
		e.logger.Log("Error parsing form - assuming light should be turned on.")
	}

	state := r.Form.Get("state")

	switch state {
	case "":
		e.logger.Log("No state provided - assuming light should be turned on.")
		break
	case "off":
		lightOn = false
		break
	case "on":
		// Do nothing - lightOn is already set to true
		break
	default:
		e.logger.Log("Invalid state provided '(" + state + ")' - assuming light should be turned on.")
	}

	if lightOn {
		e.logger.Log("Turning light on")
		_, err = e.executeCommand(e.gpioExecutable, GpioWriteCommand, tostr(e.gpioLightPin), GpioHighState)
		if err != nil {
			e.handleLightError(w)
		} else {
			w.Write([]byte("light on"))
		}
	} else {
		e.logger.Log("Turning light off")
		_, err = e.executeCommand(e.gpioExecutable, GpioWriteCommand, tostr(e.gpioLightPin), GpioLowState)
		if err != nil {
			e.handleLightError(w)
		} else {
			w.Write([]byte("light off"))
		}
	}
}

func (e Executor) handleLightError(w http.ResponseWriter) {
	e.logger.Log("Error occured while executing " + e.gpioExecutable + " " + GpioWriteCommand)
	w.Write([]byte("error - light state unchanged"))
}

func (e *Executor) executeCommand(executable string, arg ...string) (string, error) {
	e.logger.Log("executing: '" + executable + " " + strings.Join(arg, " ") + "'")
	out, err := e.osHelper.Exec(executable, arg...)
	if err != nil {
		e.logger.Log(err.Error())
	}
	return out, err
}
