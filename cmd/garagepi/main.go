package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/robdimsdale/garagepi"
	"github.com/robdimsdale/garagepi/fshelper"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/oshelper"
)

var (
	port = flag.Uint("port", 9999, "Port for server to bind to.")

	webcamHost = flag.String("webcamHost", "localhost", "Host of webcam image.")
	webcamPort = flag.Uint("webcamPort", 8080, "Port of webcam image.")

	loggingOn = flag.Bool("loggingOn", true, "Whether logging is enabled.")

	gpioExecutable = flag.String("gpioExecutable", "gpio", "Executable of gpio application.")
	gpioDoorPin    = flag.Uint("gpioDoorPin", 0, "Gpio pin of door.")
	gpioLightPin   = flag.Uint("gpioLightPin", 8, "Gpio pin of light.")
)

func main() {
	flag.Parse()

	logger := logger.NewLoggerImpl(*loggingOn)

	// The location of the 'assets' directory
	// is relative to where the compilation takes place
	// This assumes compliation happens from the root directory
	// It is also apparently relative to the fshelper package.
	fsHelper := fshelper.NewFsHelperImpl("../assets")
	osHelper := oshelper.NewOsHelperImpl(logger)
	httpHelper := garagepi.NewHttpHelperImpl()

	rtr := mux.NewRouter()

	config := garagepi.ExecutorConfig{
		WebcamHost:     *webcamHost,
		WebcamPort:     *webcamPort,
		GpioDoorPin:    *gpioDoorPin,
		GpioLightPin:   *gpioLightPin,
		GpioExecutable: *gpioExecutable,
	}

	gpio := gpio.NewGpio(osHelper, config.GpioExecutable)
	e := garagepi.NewExecutor(
		logger,
		osHelper,
		fsHelper,
		httpHelper,
		gpio,
		config)

	staticFileSystem, err := fsHelper.GetStaticFileSystem()
	if err != nil {
		panic(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFileSystem)))

	rtr.HandleFunc("/", e.HomepageHandler).Methods("GET")
	rtr.HandleFunc("/webcam", e.WebcamHandler).Methods("GET")
	rtr.HandleFunc("/toggle", e.ToggleDoorHandler).Methods("POST")
	rtr.HandleFunc("/light", e.GetLightHandler).Methods("GET")
	rtr.HandleFunc("/light", e.SetLightHandler).Methods("POST")

	http.Handle("/", rtr)
	fmt.Printf("Listening on port %d...\n", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
