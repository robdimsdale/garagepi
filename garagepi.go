package garagepi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/robdimsdale/garagepi/fshelper"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/oshelper"
)

var (
	SleepTime = 500 * time.Millisecond
)

type ExecutorConfig struct {
	WebcamHost     string
	WebcamPort     uint
	GpioDoorPin    uint
	GpioLightPin   uint
	GpioExecutable string
}

type Executor struct {
	logger         logger.Logger
	osHelper       oshelper.OsHelper
	fsHelper       fshelper.FsHelper
	httpHelper     httphelper.HttpHelper
	gpio           gpio.Gpio
	webcamUrl      string
	gpioDoorPin    uint
	gpioLightPin   uint
	gpioExecutable string
}

func NewExecutor(
	logger logger.Logger,
	osHelper oshelper.OsHelper,
	fsHelper fshelper.FsHelper,
	httpHelper httphelper.HttpHelper,
	gpio gpio.Gpio,
	config ExecutorConfig) *Executor {

	webcamUrl := fmt.Sprintf("http://%s:%d/?action=snapshot&n=", config.WebcamHost, config.WebcamPort)

	return &Executor{
		httpHelper:     httpHelper,
		logger:         logger,
		webcamUrl:      webcamUrl,
		osHelper:       osHelper,
		fsHelper:       fsHelper,
		gpio:           gpio,
		gpioDoorPin:    config.GpioDoorPin,
		gpioLightPin:   config.GpioLightPin,
		gpioExecutable: config.GpioExecutable,
	}
}

func (e Executor) HomepageHandler(w http.ResponseWriter, r *http.Request) {
	e.logger.Log(fmt.Sprintf("%s request to %v", r.Method, r.URL))
	e.handleHomepage(w, r)
}

func (e Executor) WebcamHandler(w http.ResponseWriter, r *http.Request) {
	e.handleWebcam(w, r)
}

func (e Executor) ToggleDoorHandler(w http.ResponseWriter, r *http.Request) {
	e.logger.Log(fmt.Sprintf("%s request to %v", r.Method, r.URL))
	e.handleDoorToggle(w, r)
}

func (e Executor) GetLightHandler(w http.ResponseWriter, r *http.Request) {
	e.logger.Log(fmt.Sprintf("%s request to %v", r.Method, r.URL))
	e.handleLightGet(w, r)
}

func (e Executor) SetLightHandler(w http.ResponseWriter, r *http.Request) {
	e.logger.Log(fmt.Sprintf("%s request to %v", r.Method, r.URL))
	e.handleLightSet(w, r)
}
