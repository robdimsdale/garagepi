package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/api/door"
	"github.com/robdimsdale/garagepi/api/light"
	"github.com/robdimsdale/garagepi/api/loglevel"
	"github.com/robdimsdale/garagepi/filesystem"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/middleware"
	gpos "github.com/robdimsdale/garagepi/os"
	"github.com/robdimsdale/garagepi/web/homepage"
	"github.com/robdimsdale/garagepi/web/login"
	"github.com/robdimsdale/garagepi/web/static"
	"github.com/robdimsdale/garagepi/web/webcam"
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

	logLevel = flag.String("logLevel", string(logger.LogLevelInfo), "log level: debug, info, error or fatal")

	enableHTTP  = flag.Bool("enableHTTP", true, "Enable HTTP traffic.")
	enableHTTPS = flag.Bool("enableHTTPS", false, "Enable HTTPS traffic.")
	forceHTTPS  = flag.Bool("forceHTTPS", false, "Redirect all HTTP traffic to HTTPS.")

	httpPort     = flag.Uint("httpPort", 13080, "Port on which to listen for HTTP (if enabled)")
	httpsPort    = flag.Uint("httpsPort", 13443, "Port on which to listen for HTTP (if enabled)")
	redirectPort = flag.Uint("redirectPort", 13443, "Port to which HTTP traffic is redirected (if forceHTTPS is enabled).")

	certFile = flag.String("certFile", "", "A PEM encoded certificate file.")
	keyFile  = flag.String("keyFile", "", "A PEM encoded private key file.")

	username = flag.String("username", "", "Username for HTTP authentication.")
	password = flag.String("password", "", "Password for HTTP authentication.")

	pidFile = flag.String("pidFile", "", "File to which PID is written")

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

	logger, sink, err := logger.InitializeLogger(logger.LogLevel(*logLevel))
	if err != nil {
		fmt.Printf("Failed to initialize logger\n")
		panic(err)
	}

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

	var tlsConfig *tls.Config
	if *keyFile != "" && *certFile != "" {
		var err error
		tlsConfig, err = createTLSConfig(*keyFile, *certFile)
		if err != nil {
			logger.Fatal("exiting. Failed to create tlsConfig", err)
		}
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

	webcamURL := fmt.Sprintf("%s:%d", *webcamHost, *webcamPort)
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

	loglevelHandler := loglevel.NewServer(
		logger,
		sink,
	)

	staticFileServer := http.FileServer(static.FS(false))

	rtr := mux.NewRouter()

	rtr.PathPrefix("/static/").Handler(staticFileServer)

	rtr.HandleFunc("/", hh.Handle).Methods("GET")
	rtr.HandleFunc("/webcam", wh.Handle).Methods("GET")

	s := rtr.PathPrefix("/api/v1").Subrouter()
	s.HandleFunc("/toggle", dh.HandleToggle).Methods("POST")
	s.HandleFunc("/light", lh.HandleGet).Methods("GET")
	s.HandleFunc("/light", lh.HandleSet).Methods("POST")
	s.HandleFunc("/loglevel", loglevelHandler.GetMinLevel).Methods("GET")
	s.HandleFunc("/loglevel", loglevelHandler.SetMinLevel).Methods("POST")

	rtr.HandleFunc("/login", loginHandler.LoginGET).Methods("GET")
	rtr.HandleFunc("/login", loginHandler.LoginPOST).Methods("POST")
	rtr.HandleFunc("/logout", loginHandler.LogoutPOST).Methods("POST")

	members := grouper.Members{}
	if *enableHTTPS {
		forceHTTPS := false
		httpsRunner := NewWebRunner(
			*httpsPort,
			logger,
			rtr,
			tlsConfig,
			forceHTTPS,
			*redirectPort,
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
		var tlsConfig *tls.Config // nil
		httpRunner := NewWebRunner(
			*httpPort,
			logger,
			rtr,
			tlsConfig,
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

	if *pidFile != "" {
		pid := os.Getpid()
		err = ioutil.WriteFile(*pidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
		if err != nil {
			logger.Fatal("Failed to write pid file", err, lager.Data{
				"pid":     pid,
				"pidFile": *pidFile,
			})
		}
	}
	logger.Info("garagepi started")

	err = <-process.Wait()
	if err != nil {
		logger.Error("Error running garagepi", err)
	}
}

func createTLSConfig(keyFile string, certFile string) (*tls.Config, error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

type webRunner struct {
	port      uint
	logger    lager.Logger
	handler   http.Handler
	tlsConfig *tls.Config
}

func NewWebRunner(
	port uint,
	logger lager.Logger,
	handler http.Handler,
	tlsConfig *tls.Config,
	forceHTTPS bool,
	redirectPort uint,
	username string,
	password string,
	cookieHandler *securecookie.SecureCookie,
) ifrit.Runner {

	m := middleware.Chain{
		middleware.NewPanicRecovery(logger),
		middleware.NewLogger(logger),
	}

	if forceHTTPS {
		m = append(m, middleware.NewHTTPSEnforcer(redirectPort))
	} else if username != "" && password != "" {
		m = append(m, middleware.NewAuth(username, password, logger, cookieHandler))
	}

	return &webRunner{
		port:      port,
		logger:    logger,
		handler:   m.Wrap(handler),
		tlsConfig: tlsConfig,
	}
}

func (r webRunner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var listener net.Listener
	var err error

	if r.tlsConfig == nil {
		r.logger.Debug("listening for TCP", lager.Data{"port": r.port})
		listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.port))
	} else {
		r.logger.Debug("listening for TLS", lager.Data{"port": r.port})
		listener, err = tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.port), r.tlsConfig)
	}

	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		err := http.Serve(listener, r.handler)
		if err != nil {
			errChan <- err
		}
	}()

	close(ready)

	if r.tlsConfig == nil {
		r.logger.Info("HTTP server listening on port", lager.Data{"port": r.port})
	} else {
		r.logger.Info("HTTPS server listening on port", lager.Data{"port": r.port})
	}

	select {
	case <-signals:
		return listener.Close()
	case err := <-errChan:
		return err
	}
}
