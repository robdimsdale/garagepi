package main

import (
	"flag"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/robdimsdale/garage-pi/garagepi"
	"github.com/robdimsdale/garage-pi/logger"
	"github.com/robdimsdale/garage-pi/oshelper"
)

var (
	port = flag.String("port", "9999", "Port for server to bind to.")

	webcamHost = flag.String("webcamHost", "localhost", "Host of webcam image.")
	webcamPort = flag.String("webcamPort", "8080", "Port of webcam image.")
)

func main() {
	flag.Parse()

	osHelper := oshelper.NewOsHelperImpl()
	staticFilesystem := rice.MustFindBox("./assets/static").HTTPBox()
	templatesFilesystem := rice.MustFindBox("./assets/templates").HTTPBox()

	loggingOn := true
	l := logger.NewLoggerImpl(loggingOn)

	rtr := mux.NewRouter()

	e := garagepi.NewExecutor(
		l,
		osHelper,
		staticFilesystem,
		templatesFilesystem,
		*webcamHost,
		*webcamPort)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFilesystem)))

	rtr.HandleFunc("/", e.HomepageHandler).Methods("GET")
	rtr.HandleFunc("/webcam", e.WebcamHandler).Methods("GET")
	rtr.HandleFunc("/toggle", e.ToggleDoorHandler).Methods("POST")
	rtr.HandleFunc("/start-camera", e.StartCameraHandler).Methods("POST")
	rtr.HandleFunc("/stop-camera", e.StopCameraHandler).Methods("POST")

	http.Handle("/", rtr)
	l.Log("Listening on port " + *port + "...")
	http.ListenAndServe(":"+*port, nil)
}
