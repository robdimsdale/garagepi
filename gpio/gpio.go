package gpio

import (
	"strconv"

	"github.com/robdimsdale/garagepi/logger"
	"github.com/robdimsdale/garagepi/oshelper"
)

const gpioReadCommand = "read"
const gpioWriteCommand = "write"

type Gpio interface {
	Read(pin uint) (string, error)
	Write(pin uint, state string) error
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

func (g gpio) Write(pin uint, state string) error {
	args := []string{gpioWriteCommand, tostr(pin), state}
	_, err := g.osHelper.Exec(g.gpioExecutable, args...)
	return err
}

func tostr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
