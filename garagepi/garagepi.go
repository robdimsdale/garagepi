package garagepi

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/robdimsdale/garage-pi/httphelper"
	"github.com/robdimsdale/garage-pi/logger"
	"github.com/robdimsdale/garage-pi/oshelper"
)

type Executor struct {
	l          logger.Logger
	osHelper   oshelper.OsHelper
	httpHelper httphelper.HttpHelper
	webcamUrl  string
}

func NewExecutor(
	l logger.Logger,
	httpHelper httphelper.HttpHelper,
	osHelper oshelper.OsHelper,
	webcamHost string,
	webcamPort string) *Executor {

	webcamUrl := "http://" + webcamHost + ":" + webcamPort + "/?action=snapshot&n="

	return &Executor{
		httpHelper: httpHelper,
		l:          l,
		webcamUrl:  webcamUrl,
		osHelper:   osHelper,
	}
}

func (e *Executor) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	e.l.Log("homepage")
	bytes, err := e.osHelper.GetHomepageTemplateContents()
	if err != nil {
		e.l.Log("Error reading homepage template: " + err.Error())
		panic(err)
	}
	w.Write(bytes)
}

func (e *Executor) WebcamHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := e.httpHelper.Get(e.webcamUrl + r.Form.Get("n"))
	if err != nil {
		e.l.Log("Error getting image: " + err.Error())
		if resp == nil {
			return
		}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.l.Log("Error closing image request: " + err.Error())
	}
	w.Write(body)
}

func (e *Executor) ToggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("gpio", "write", "0", "1")
	e.osHelper.Sleep(500 * time.Millisecond)
	e.executeCommand("gpio", "write", "0", "0")

	e.httpHelper.RedirectToHomepage(w, r)
}

func (e *Executor) executeCommand(executable string, arg ...string) string {
	e.l.Log("executing: '" + executable + " " + strings.Join(arg, " ") + "'")
	out, err := e.osHelper.Exec(executable, arg...)
	if err != nil {
		e.l.Log("ERROR: " + err.Error())
	}
	return out
}

func (e *Executor) StartCameraHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("/etc/init.d/garagestreamer", "start")
	e.httpHelper.RedirectToHomepage(w, r)
}

func (e *Executor) StopCameraHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("/etc/init.d/garagestreamer", "stop")
	e.httpHelper.RedirectToHomepage(w, r)
}
