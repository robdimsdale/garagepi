package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robdimsdale/garagepi"
)

var (
	port = flag.Uint("port", 9999, "Port for server to bind to.")

	webcamHost = flag.String("webcamHost", "localhost", "Host of webcam image.")
	webcamPort = flag.Uint("webcamPort", 8080, "Port of webcam image.")

	loggingOn = flag.Bool("loggingOn", true, "Whether logging is enabled.")
)

func main() {
	flag.Parse()

	logger := garagepi.NewLoggerImpl(*loggingOn)

	// The location of the 'assets' directory
	// is relative to where the compilation takes place
	// This assumes compliation happens from the root directory
	osHelper := garagepi.NewOsHelperImpl(logger, "assets")
	httpHelper := garagepi.NewHttpHelperImpl()

	rtr := mux.NewRouter()

	e := garagepi.NewExecutor(
		logger,
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
	fmt.Printf("Listening on port %d...\n", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
