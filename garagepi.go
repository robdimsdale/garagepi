package garagepi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/robdimsdale/garagepi/fshelper"
	"github.com/robdimsdale/garagepi/gpio"
	"github.com/robdimsdale/garagepi/httphelper"
	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/oshelper"
)

var (
	SleepTime        = 500 * time.Millisecond
	GpioReadCommand  = "read"
	GpioWriteCommand = "write"
	GpioLowState     = "0"
	GpioHighState    = "1"
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
	g              gpio.Gpio
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
	g gpio.Gpio,
	config ExecutorConfig) *Executor {

	webcamUrl := fmt.Sprintf("http://%s:%d/?action=snapshot&n=", config.WebcamHost, config.WebcamPort)

	return &Executor{
		httpHelper:     httpHelper,
		logger:         logger,
		webcamUrl:      webcamUrl,
		osHelper:       osHelper,
		fsHelper:       fsHelper,
		g:              g,
		gpioDoorPin:    config.GpioDoorPin,
		gpioLightPin:   config.GpioLightPin,
		gpioExecutable: config.GpioExecutable,
	}
}

func onOffStringToStateNumber(onOff string) (string, error) {
	switch onOff {
	case "on":
		return GpioHighState, nil
	case "off":
		return GpioLowState, nil
	default:
		return "", errors.New(fmt.Sprintf("Unrecognized state: %s", onOff))
	}
}

func tostr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
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

func (e Executor) executeCommand(executable string, arg ...string) (string, error) {
	out, err := e.osHelper.Exec(executable, arg...)
	if err != nil {
		e.logger.Log(err.Error())
	}
	return out, err
}
