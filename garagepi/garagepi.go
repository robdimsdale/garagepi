package garagepi

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Executor struct {
	rtr      *mux.Router
	osHelper OsHelper
}

type OsHelper interface {
	Exec(executable string, arg ...string) string
}

// type osHelperImpl struct {
// }

func NewExecutor(helper OsHelper) *Executor {
	e := new(Executor)
	e.rtr = mux.NewRouter()
	e.osHelper = helper
	return e
}

func (e *Executor) homepageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("homepage")
	http.ServeFile(w, r, "./templates/homepage.html")
}

func (e *Executor) toggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("bash", "gpio-toggle.sh")
	http.Redirect(w, r, "/", 303)
}

func (e *Executor) executeCommand(executable string, arg ...string) string {
	logStatement := append([]string{executable}, arg...)
	log.Println("executing", logStatement)
	return e.osHelper.Exec(executable, arg...)
}

func (e *Executor) startCameraHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("bash", "start-camera.sh")
	http.Redirect(w, r, "/", 303)
}

func (e *Executor) stopCameraHandler(w http.ResponseWriter, r *http.Request) {
	e.executeCommand("bash", "stop-camera.sh")
	http.Redirect(w, r, "/", 303)
}

func (e *Executor) ServeForever(port string) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	e.rtr.HandleFunc("/", e.homepageHandler).Methods("GET")
	e.rtr.HandleFunc("/toggle", e.toggleDoorHandler).Methods("POST")
	e.rtr.HandleFunc("/start-camera", e.startCameraHandler).Methods("POST")
	e.rtr.HandleFunc("/stop-camera", e.stopCameraHandler).Methods("POST")

	http.Handle("/", e.rtr)
	log.Println("Listening on port " + port + "...")
	http.ListenAndServe(":"+port, nil)
}
