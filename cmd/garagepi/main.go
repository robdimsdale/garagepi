package main

import (
	"flag"
	"net/http"

	"github.com/gorilla/mux"
	garagepi "github.com/robdimsdale/garage-pi"
)

var (
	port = flag.String("port", "9999", "Port for server to bind to.")

	webcamHost = flag.String("webcamHost", "localhost", "Host of webcam image.")
	webcamPort = flag.String("webcamPort", "8080", "Port of webcam image.")
)

func main() {
	flag.Parse()

	loggingOn := true
	l := garagepi.NewLoggerImpl(loggingOn)

	osHelper := garagepi.NewOsHelperImpl(l, "../assets")
	httpHelper := garagepi.NewHttpHelperImpl()

	rtr := mux.NewRouter()

	e := garagepi.NewExecutor(
		l,
		httpHelper,
		osHelper,
		*webcamHost,
		*webcamPort)

	staticFileSystem, err := garagepi.GetStaticFileSystem()
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
