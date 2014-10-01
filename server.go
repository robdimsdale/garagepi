package main

import (
	"flag"
	"os/exec"

	"github.com/GeertJohan/go.rice"
	"github.com/robdimsdale/garage-pi/garagepi"
)

var (
	port = flag.String("port", "9999", "Port for server to bind to.")

	webcamHost = flag.String("webcamHost", "localhost", "Host of webcam image.")
	webcamPort = flag.String("webcamPort", "8080", "Port of webcam image.")
)

type osHelperImpl struct {
}

func (h *osHelperImpl) Exec(executable string, arg ...string) (string, error) {
	out, err := exec.Command(executable, arg...).CombinedOutput()
	return string(out), err
}

func main() {

	flag.Parse()

	osHelper := new(osHelperImpl)
	staticFilesystem := rice.MustFindBox("./assets/static").HTTPBox()
	templatesFilesystem := rice.MustFindBox("./assets/templates").HTTPBox()
	e := garagepi.NewExecutor(
		osHelper,
		staticFilesystem,
		templatesFilesystem,
		*webcamHost,
		*webcamPort)
	e.ServeForever(*port)
}
