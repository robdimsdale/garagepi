package garagepi

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"io/ioutil"
)

type Executor struct {
	webcamHost          string
	webcamPort          string
	rtr                 *mux.Router
	osHelper            OsHelper
	staticFilesystem    http.FileSystem
	templatesFilesystem http.FileSystem
}

type OsHelper interface {
	Exec(executable string, arg ...string) (string, error)
}

func NewExecutor(
	helper OsHelper,
	staticFilesystem http.FileSystem,
	templatesFilesystem http.FileSystem,
	webcamHost string,
	webcamPort string) *Executor {

	e := new(Executor)
	e.webcamHost = webcamHost
	e.webcamPort = webcamPort
	e.rtr = mux.NewRouter()
	e.osHelper = helper
	e.staticFilesystem = staticFilesystem
	e.templatesFilesystem = templatesFilesystem
	return e
}

func (e *Executor) homepageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("homepage")
	buf := bytes.NewBuffer(nil)
	f, err := e.templatesFilesystem.Open("homepage.html")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(buf, f)
	if err != nil {
		panic(err)
	}
	f.Close()
	w.Write(buf.Bytes())
}

func (e *Executor) webcamHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://" + e.webcamHost + ":" + e.webcamPort + "/?action=snapshot&n=" + r.Form.Get("n"))
	if err != nil {
		log.Println("Error getting image: " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error closing image request: " + err.Error())
	}
	w.Write(body)
}

func (e *Executor) toggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("gpio", "write", "0", "1")
	sleepDuration, _ := time.ParseDuration("500ms")
	log.Println("sleeping for", sleepDuration)
	time.Sleep(sleepDuration)
	e.executeCommand("gpio", "write", "0", "0")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (e *Executor) executeCommand(executable string, arg ...string) string {
	logStatement := append([]string{executable}, arg...)
	log.Println("executing", logStatement)
	out, err := e.osHelper.Exec(executable, arg...)
	if err != nil {
		log.Println("ERROR:", err)
	}
	return out
}

func (e *Executor) startCameraHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("/etc/init.d/garagestreamer", "start")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (e *Executor) stopCameraHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("/etc/init.d/garagestreamer", "stop")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (e *Executor) ServeForever(port string) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(e.staticFilesystem)))

	e.rtr.HandleFunc("/", e.homepageHandler).Methods("GET")
	e.rtr.HandleFunc("/webcam", e.webcamHandler).Methods("GET")
	e.rtr.HandleFunc("/toggle", e.toggleDoorHandler).Methods("POST")
	e.rtr.HandleFunc("/start-camera", e.startCameraHandler).Methods("POST")
	e.rtr.HandleFunc("/stop-camera", e.stopCameraHandler).Methods("POST")

	http.Handle("/", e.rtr)
	log.Println("Listening on port " + port + "...")
	http.ListenAndServe(":"+port, nil)
}

func (e *Executor) Kill() {
	os.Exit(0)
}
