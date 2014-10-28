package main

import (
	"flag"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robdimsdale/garage-pi/garagepi"
	"github.com/robdimsdale/garage-pi/httphelper"
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

	loggingOn := true
	l := logger.NewLoggerImpl(loggingOn)

	osHelper := oshelper.NewOsHelperImpl(l, "../assets")
	httpHelper := httphelper.NewHttpHelperImpl()

	rtr := mux.NewRouter()

	e := garagepi.NewExecutor(
		l,
		httpHelper,
		osHelper,
		*webcamHost,
		*webcamPort)

	staticFileSystem, err := osHelper.GetStaticFileSystem()
	if err != nil {
		panic(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFileSystem)))

	rtr.HandleFunc("/", e.HomepageHandler).Methods("GET")
	rtr.HandleFunc("/webcam", e.WebcamHandler).Methods("GET")
	rtr.HandleFunc("/toggle", e.ToggleDoorHandler).Methods("POST")

	http.Handle("/", rtr)
	l.Log("Listening on port " + *port + "...")
	http.ListenAndServe(":"+*port, nil)
}
