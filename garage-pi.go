package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

var serverDir string

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("homepage")
	http.ServeFile(w, r, "./templates/homepage.html")
}

func toggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	executeCommand("bash", serverDir+"gpio-toggle.sh")
	http.Redirect(w, r, "/", 303)
}

func executeCommand(executable string, arg ...string) string {
	logStatement := append([]string{executable}, arg...)
	log.Println("executing", logStatement)
	out, _ := exec.Command(executable, arg...).Output()
	return string(out)
}

func startCameraHandler(w http.ResponseWriter, r *http.Request) {
	executeCommand("bash", serverDir+"start-camera.sh")
	http.Redirect(w, r, "/", 303)
}

func stopCameraHandler(w http.ResponseWriter, r *http.Request) {
	executeCommand("bash", serverDir+"stop-camera.sh")
	http.Redirect(w, r, "/", 303)
}

func main() {
	serverDir := getServerDir()
	log.Println("Using serverDir:" + serverDir)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	rtr := mux.NewRouter()

	rtr.HandleFunc("/", homepageHandler).Methods("GET")
	rtr.HandleFunc("/toggle", toggleDoorHandler).Methods("POST")
	rtr.HandleFunc("/start-camera", startCameraHandler).Methods("POST")
	rtr.HandleFunc("/stop-camera", stopCameraHandler).Methods("POST")

	http.Handle("/", rtr)
	port := 9999
	portAsString := strconv.Itoa(port)
	log.Println("Listening on port " + portAsString + "...")
	http.ListenAndServe(":"+portAsString, nil)
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
