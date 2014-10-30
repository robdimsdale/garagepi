package garagepi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	SleepTime        = 500 * time.Millisecond
	GpioPin          = "0"
	GpioLightPin     = "8"
	GpioExecutable   = "gpio"
	GpioWriteCommand = "write"
	GpioLowState     = "0"
	GpioHighState    = "1"
)

type Executor struct {
	logger     Logger
	osHelper   OsHelper
	fsHelper   FsHelper
	httpHelper HttpHelper
	webcamUrl  string
}

func NewExecutor(
	logger Logger,
	httpHelper HttpHelper,
	osHelper OsHelper,
	fsHelper FsHelper,
	webcamHost string,
	webcamPort uint) *Executor {

	webcamUrl := fmt.Sprintf("http://%s:%d/?action=snapshot&n=", webcamHost, webcamPort)

	return &Executor{
		httpHelper: httpHelper,
		logger:     logger,
		webcamUrl:  webcamUrl,
		osHelper:   osHelper,
		fsHelper:   fsHelper,
	}
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
	_, err := e.executeCommand(GpioExecutable, GpioWriteCommand, GpioPin, GpioHighState)
	if err != nil {
		e.logger.Log("Error occured while executing " + GpioWriteCommand + " - skipping sleep and further executions")
	} else {
		e.osHelper.Sleep(SleepTime)
		e.executeCommand(GpioExecutable, GpioWriteCommand, GpioPin, GpioLowState)
	}

	e.httpHelper.RedirectToHomepage(w, r)
}

func (e *Executor) LightHandler(w http.ResponseWriter, r *http.Request) {
	lightOn := true
	err := r.ParseForm()
	if err != nil {
		e.logger.Log("Error parsing form - assuming light should be turned on.")
	}
	if r.Form.Get("state") == "off" {
		lightOn = false
	}

	if lightOn {
		_, err = e.executeCommand(GpioExecutable, GpioWriteCommand, GpioLightPin, GpioHighState)
	} else {
		_, err = e.executeCommand(GpioExecutable, GpioWriteCommand, GpioLightPin, GpioLowState)
	}

	if err != nil {
		e.logger.Log("Error occured while executing " + GpioWriteCommand + " - skipping sleep and further executions")
	}

	e.httpHelper.RedirectToHomepage(w, r)
}

func (e *Executor) executeCommand(executable string, arg ...string) (string, error) {
	e.logger.Log("executing: '" + executable + " " + strings.Join(arg, " ") + "'")
	out, err := e.osHelper.Exec(executable, arg...)
	if err != nil {
		e.logger.Log(err.Error())
	}
	return out, err
}
