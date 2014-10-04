package garagepi

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/robdimsdale/garage-pi/logger"
	"github.com/robdimsdale/garage-pi/oshelper"
)

type Executor struct {
	l         logger.Logger
	webcamUrl string
	osHelper  oshelper.OsHelper
}

func NewExecutor(
	l logger.Logger,
	helper oshelper.OsHelper,
	webcamHost string,
	webcamPort string) *Executor {

	webcamUrl := "http://" + webcamHost + ":" + webcamPort + "/?action=snapshot&n="

	return &Executor{
		l:         l,
		webcamUrl: webcamUrl,
		osHelper:  helper,
	}
}

func (e *Executor) panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (e *Executor) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	e.l.Log("homepage")
	f, err := e.osHelper.GetHomepageTemplate()
	e.panicOnErr(err)
	defer f.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, f)
	e.panicOnErr(err)

	w.Write(buf.Bytes())
}

func (e *Executor) WebcamHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(e.webcamUrl + r.Form.Get("n"))
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
	sleepDuration, _ := time.ParseDuration("500ms")
	e.l.Log("sleeping for " + sleepDuration.String())
	time.Sleep(sleepDuration)
	e.executeCommand("gpio", "write", "0", "0")

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (e *Executor) StopCameraHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("/etc/init.d/garagestreamer", "stop")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
