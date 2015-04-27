package gpio

import (
	"strconv"

	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/oshelper"
)

const gpioReadCommand = "read"
const gpioWriteCommand = "write"
const gpioLowState = "0"
const gpioHighState = "1"

type Gpio interface {
	Read(pin uint) (string, error)
	WriteLow(pin uint) error
	WriteHigh(pin uint) error
}

type gpio struct {
	osHelper       oshelper.OsHelper
	logger         logger.Logger
	gpioExecutable string
}

func NewGpio(
	osHelper oshelper.OsHelper,
	logger logger.Logger,
	gpioExecutable string,
) Gpio {
	return &gpio{
		osHelper:       osHelper,
		logger:         logger,
		gpioExecutable: gpioExecutable,
	}
}

func (g gpio) Read(pin uint) (string, error) {
	args := []string{gpioReadCommand, tostr(pin)}
	return g.osHelper.Exec(g.gpioExecutable, args...)
}

func (g gpio) WriteLow(pin uint) error {
	args := []string{gpioWriteCommand, tostr(pin), gpioLowState}
	_, err := g.osHelper.Exec(g.gpioExecutable, args...)
	return err
}

func (g gpio) WriteHigh(pin uint) error {
	args := []string{gpioWriteCommand, tostr(pin), gpioHighState}
	_, err := g.osHelper.Exec(g.gpioExecutable, args...)
	return err
}

func tostr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
