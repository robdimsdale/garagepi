package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/robdimsdale/garage-pi/garagepi"
)

var defaultPort = "9999"

type osHelperImpl struct {
}

func (h *osHelperImpl) Exec(executable string, arg ...string) string {
	out, _ := exec.Command(executable, arg...).Output()
	return string(out)
}

func main() {
	serverDir := getServerDir()
	log.Println("Running from: " + serverDir)

	port := flag.String("port", defaultPort, "help message for flagname")
	flag.Parse()

	osHelper := new(osHelperImpl)
	e := garagepi.NewExecutor(osHelper)
	e.ServeForever(*port)
}

func getServerDir() string {
	osDir := os.Getenv("SERVER_DIR")
	if osDir == "" {
		log.Println("SERVER_DIR env variable not found")
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		osDir = dir
	}
	return osDir
}
