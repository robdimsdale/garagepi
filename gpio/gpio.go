package gpio

import (
	"fmt"
	"strconv"

	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/oshelper"
	"github.com/stianeikeland/go-rpio"
)

type Gpio interface {
	Read(pin uint) (string, error)
	WriteLow(pin uint) error
	WriteHigh(pin uint) error
}

type gpio struct {
	osHelper oshelper.OsHelper
	logger   logger.Logger
}

func NewGpio(
	osHelper oshelper.OsHelper,
	logger logger.Logger,
) Gpio {
	return &gpio{
		osHelper: osHelper,
		logger:   logger,
	}
}

func (g gpio) Read(pin uint) (string, error) {
	g.logger.Log(fmt.Sprintf("Reading from pin: %d", pin))

	rpin := rpio.Pin(pin)

	err := rpio.Open()
	if err != nil {
		return "", err
	}
	defer rpio.Close()

	state := rpin.Read()
	return fmt.Sprintf("%v", state), err
}

func (g gpio) WriteLow(pin uint) error {
	g.logger.Log(fmt.Sprintf("Writing low to pin: %d", pin))

	rpin := rpio.Pin(pin)

	err := rpio.Open()
	if err != nil {
		return err
	}
	defer rpio.Close()

	rpin.Output()
	rpin.Low()
	return nil
}

func (g gpio) WriteHigh(pin uint) error {
	g.logger.Log(fmt.Sprintf("Writing high to pin: %d", pin))

	rpin := rpio.Pin(pin)

	err := rpio.Open()
	if err != nil {
		return err
	}
	defer rpio.Close()

	rpin.Output()
	rpin.High()
	return nil
}

func tostr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
