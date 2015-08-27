package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/door"
	"github.com/robdimsdale/garagepi/filesystem"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/handler"
	"github.com/robdimsdale/garagepi/homepage"
	"github.com/robdimsdale/garagepi/light"
	"github.com/robdimsdale/garagepi/logger"
	gpos "github.com/robdimsdale/garagepi/os"
	"github.com/robdimsdale/garagepi/static"
	"github.com/robdimsdale/garagepi/webcam"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string

	webcamHost = flag.String("webcamHost", "localhost", "Host of webcam image.")
	webcamPort = flag.Uint("webcamPort", 8080, "Port of webcam image.")

	gpioDoorPin  = flag.Uint("gpioDoorPin", 17, "Gpio pin of door.")
	gpioLightPin = flag.Uint("gpioLightPin", 2, "Gpio pin of light.")

	logLevel = flag.String("logLevel", string(logger.INFO), "log level: debug, info, error or fatal")

	enableHTTP  = flag.Bool("enableHTTP", true, "Enable HTTP traffic.")
	enableHTTPS = flag.Bool("enableHTTPS", false, "Enable HTTPS traffic.")
	forceHTTPS  = flag.Bool("forceHTTPS", false, "Redirect all HTTP traffic to HTTPS.")

	httpPort  = flag.Uint("httpPort", 13080, "Port on which to listen for HTTP (if enabled)")
	httpsPort = flag.Uint("httpsPort", 13433, "Port on which to listen for HTTP (if enabled)")

	certFile = flag.String("certFile", "", "A PEM encoded certificate file.")
	keyFile  = flag.String("keyFile", "", "A PEM encoded private key file.")
	caFile   = flag.String("caFile", "", "A PEM encoded CA's certificate file.")

	username = flag.String("username", "", "Username for HTTP authentication.")
	password = flag.String("password", "", "Password for HTTP authentication.")

	dev = flag.Bool("dev", false, "Development mode; do not require username/password")
)

func main() {
	if version == "" {
		version = "dev"
	}

	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "version" || arg == "-v" || arg == "--version" {
			fmt.Printf("%s\n", version)
			os.Exit(0)
		}
	}

	flag.Parse()

	logger := logger.InitializeLogger(*logLevel)
	logger.Info("garagepi starting", lager.Data{"version": version})
	logger.Debug("flags", lager.Data{
		"enableHTTP":  enableHTTP,
		"enableHTTPS": enableHTTPS,
		"forceHTTPS":  forceHTTPS,
	})

	if !(*enableHTTP || *enableHTTPS) {
		logger.Fatal("exiting", fmt.Errorf("at least one of enableHTTP and enableHTTPS must be true"))
	}

	if *enableHTTPS {
		if *keyFile == "" {
			logger.Fatal("exiting", fmt.Errorf("keyFile must be provided if enableHTTPS is true"))
		}

		if *certFile == "" {
			logger.Fatal("exiting", fmt.Errorf("certFile must be provided if enableHTTPS is true"))
		}
	}

	if *forceHTTPS && !(*enableHTTP && *enableHTTPS) {
		logger.Fatal("exiting", fmt.Errorf("enableHTTP must be enabled if forceHTTPS is true"))
	}

	if !*dev && (*username == "" || *password == "") {
		logger.Fatal("exiting", fmt.Errorf("must specify -username and -password or turn on dev mode"))
	}

	// The location of the 'assets' directory
	// is relative to where the compilation takes place
	// This assumes compliation happens from the root directory
	// It is also apparently relative to the filesystem package.
	fsHelper := filesystem.NewFileSystemHelper()
	osHelper := gpos.NewOSHelper(logger)

	webcamURL := fmt.Sprintf("http://%s:%d/?action=snapshot&n=", *webcamHost, *webcamPort)
	wh := webcam.NewHandler(
		logger,
		webcamURL,
	)

	gpio := gpio.NewGpio(osHelper, logger)

	lh := light.NewHandler(
		logger,
		gpio,
		*gpioLightPin,
	)

	hh := homepage.NewHandler(
		logger,
		fsHelper,
		lh,
	)

	dh := door.NewHandler(
		logger,
		osHelper,
		gpio,
		*gpioDoorPin)

	// staticFileSystem, err := fsHelper.GetStaticFileSystem()
	staticFileSystem := static.FS(false)

	staticFileServer := http.FileServer(staticFileSystem)
	// strippedStaticFileServer := http.StripPrefix("/static/", staticFileServer)

	rtr := mux.NewRouter()

	// rtr.PathPrefix("/static/").Handler(strippedStaticFileServer)
	rtr.PathPrefix("/static/").Handler(staticFileServer)
	rtr.HandleFunc("/", hh.Handle).Methods("GET")
	rtr.HandleFunc("/webcam", wh.Handle).Methods("GET")
	rtr.HandleFunc("/toggle", dh.HandleToggle).Methods("POST")
	rtr.HandleFunc("/light", lh.HandleGet).Methods("GET")
	rtr.HandleFunc("/light", lh.HandleSet).Methods("POST")

	members := grouper.Members{}
	if *enableHTTPS {
		httpsRunner := handler.NewHTTPSRunner(
			*httpsPort,
			logger,
			rtr,
			*keyFile,
			*certFile,
			*caFile,
			*username,
			*password,
		)

		members = append(members, grouper.Member{
			Name:   "https",
			Runner: httpsRunner,
		})
	}

	if *enableHTTP {
		httpRunner := handler.NewHTTPRunner(
			*httpPort,
			logger,
			rtr,
			*forceHTTPS,
			*httpsPort,
			*username,
			*password,
		)
		members = append(members, grouper.Member{
			Name:   "http",
			Runner: httpRunner,
		})
	}

	group := grouper.NewParallel(os.Kill, members)
	process := ifrit.Invoke(group)

	logger.Info("garagepi started")

	err := <-process.Wait()
	if err != nil {
		logger.Error("Error running garagepi", err)
	}
}
