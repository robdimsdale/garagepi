package main

import (
	"flag"
	"os/exec"

	"github.com/GeertJohan/go.rice"
	"github.com/robdimsdale/garage-pi/garagepi"
)

var defaultPort = "9999"

type osHelperImpl struct {
}

func (h *osHelperImpl) Exec(executable string, arg ...string) (string, error) {
	out, err := exec.Command(executable, arg...).CombinedOutput()
	return string(out), err
}

func main() {
	port := flag.String("port", defaultPort, "help message for flagname")
	flag.Parse()

	osHelper := new(osHelperImpl)
	staticFilesystem := rice.MustFindBox("./assets/static").HTTPBox()
	templatesFilesystem := rice.MustFindBox("./assets/templates").HTTPBox()
	e := garagepi.NewExecutor(osHelper, staticFilesystem, templatesFilesystem)
	e.ServeForever(*port)
}
