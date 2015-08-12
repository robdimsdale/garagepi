package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/door"
	"github.com/robdimsdale/garagepi/fshelper"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/homepage"
	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/light"
	"github.com/robdimsdale/garagepi/oshelper"
	"github.com/robdimsdale/garagepi/webcam"
	"github.com/tedsuo/ifrit"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	ERROR = "error"
	FATAL = "fatal"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string

	port = flag.Uint("port", 9999, "Port for server to bind to.")

	webcamHost = flag.String("webcamHost", "localhost", "Host of webcam image.")
	webcamPort = flag.Uint("webcamPort", 8080, "Port of webcam image.")

	gpioDoorPin  = flag.Uint("gpioDoorPin", 17, "Gpio pin of door.")
	gpioLightPin = flag.Uint("gpioLightPin", 2, "Gpio pin of light.")

	logLevel = flag.String("logLevel", string(INFO), "log level: debug, info, error or fatal")
)

func main() {
	if version == "" {
		version = "dev"
	}

	flag.Parse()

	logger := initializeLogger()
	logger.Info("garagepi starting", lager.Data{"version": version})

	// The location of the 'assets' directory
	// is relative to where the compilation takes place
	// This assumes compliation happens from the root directory
	// It is also apparently relative to the fshelper package.
	fsHelper := fshelper.NewFsHelperImpl("../assets")
	osHelper := oshelper.NewOsHelperImpl(logger)
	httpHelper := httphelper.NewHttpHelperImpl()

	rtr := mux.NewRouter()

	wh := webcam.NewHandler(
		logger,
		httpHelper,
		*webcamHost,
		*webcamPort,
	)

	gpio := gpio.NewGpio(osHelper, logger)

	lh := light.NewHandler(
		logger,
		httpHelper,
		gpio,
		*gpioLightPin,
	)

	hh := homepage.NewHandler(
		logger,
		httpHelper,
		fsHelper,
		lh,
	)

	dh := door.NewHandler(
		logger,
		httpHelper,
		osHelper,
		gpio,
		*gpioDoorPin)

	staticFileSystem, err := fsHelper.GetStaticFileSystem()
	if err != nil {
		panic(err)
	}

	staticFileServer := http.FileServer(staticFileSystem)
	strippedStaticFileServer := http.StripPrefix("/static/", staticFileServer)

	rtr.PathPrefix("/static/").Handler(strippedStaticFileServer)
	rtr.HandleFunc("/", hh.Handle).Methods("GET")
	rtr.HandleFunc("/webcam", wh.Handle).Methods("GET")
	rtr.HandleFunc("/toggle", dh.HandleToggle).Methods("POST")
	rtr.HandleFunc("/light", lh.HandleGet).Methods("GET")
	rtr.HandleFunc("/light", lh.HandleSet).Methods("POST")

	http.Handle("/", rtr)

	runner := runner{
		port:    *port,
		logger:  logger,
		handler: rtr,
	}

	process := ifrit.Invoke(runner)

	logger.Info("garagepi started")

	err = <-process.Wait()
	if err != nil {
		logger.Error("Error running garagepi", err)
	}
}

func initializeLogger() lager.Logger {
	var minLagerLogLevel lager.LogLevel
	switch *logLevel {
	case DEBUG:
		minLagerLogLevel = lager.DEBUG
	case INFO:
		minLagerLogLevel = lager.INFO
	case ERROR:
		minLagerLogLevel = lager.ERROR
	case FATAL:
		minLagerLogLevel = lager.FATAL
	default:
		panic(fmt.Errorf("unknown log level: %s", *logLevel))
	}

	logger := lager.NewLogger("garagepi")

	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), minLagerLogLevel)
	logger.RegisterSink(sink)

	return logger
}

type runner struct {
	port    uint
	logger  lager.Logger
	handler http.Handler
}

func (r runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.port))
	if err != nil {
		return err
	} else {
		r.logger.Info("web server listening on port", lager.Data{"port": r.port})
	}

	errChan := make(chan error)
	go func() {
		err := http.Serve(listener, r.handler)
		if err != nil {
			errChan <- err
		}
	}()

	close(ready)

	select {
	case <-signals:
		return listener.Close()
	case err := <-errChan:
		return err
	}
}
