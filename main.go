package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/door"
	"github.com/robdimsdale/garagepi/filesystem"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/handler"
	"github.com/robdimsdale/garagepi/homepage"
	"github.com/robdimsdale/garagepi/light"
	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/login"
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

	httpPort     = flag.Uint("httpPort", 13080, "Port on which to listen for HTTP (if enabled)")
	httpsPort    = flag.Uint("httpsPort", 13433, "Port on which to listen for HTTP (if enabled)")
	redirectPort = flag.Uint("redirectPort", 13443, "Port to which HTTP traffic is redirected (if forceHTTPS is enabled).")

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

	cookieHandler := securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)

	templates, err := filesystem.LoadTemplates()
	if err != nil {
		logger.Fatal("exiting", err)
	}

	osHelper := gpos.NewOSHelper(logger)

	loginHandler := login.NewHandler(
		logger,
		templates,
		cookieHandler,
	)

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
		templates,
		lh,
		loginHandler,
	)

	dh := door.NewHandler(
		logger,
		osHelper,
		gpio,
		*gpioDoorPin)

	staticFileServer := http.FileServer(static.FS(false))

	rtr := mux.NewRouter()

	rtr.PathPrefix("/static/").Handler(staticFileServer)
	rtr.HandleFunc("/", hh.Handle).Methods("GET")
	rtr.HandleFunc("/webcam", wh.Handle).Methods("GET")
	rtr.HandleFunc("/toggle", dh.HandleToggle).Methods("POST")
	rtr.HandleFunc("/light", lh.HandleGet).Methods("GET")
	rtr.HandleFunc("/light", lh.HandleSet).Methods("POST")

	rtr.HandleFunc("/login", loginHandler.LoginGET).Methods("GET")
	rtr.HandleFunc("/login", loginHandler.LoginPOST).Methods("POST")
	rtr.HandleFunc("/logout", loginHandler.LogoutPOST).Methods("POST")

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
			cookieHandler,
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
			*redirectPort,
			*username,
			*password,
			cookieHandler,
		)
		members = append(members, grouper.Member{
			Name:   "http",
			Runner: httpRunner,
		})
	}

	group := grouper.NewParallel(os.Kill, members)
	process := ifrit.Invoke(group)

	logger.Info("garagepi started")

	err = <-process.Wait()
	if err != nil {
		logger.Error("Error running garagepi", err)
	}
}
