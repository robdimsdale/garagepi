package gpio

import (
	"strconv"

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
	gpioExecutable string
}

func NewGpio(
	osHelper oshelper.OsHelper,
	gpioExecutable string,
) Gpio {
	return &gpio{
		osHelper:       osHelper,
		gpioExecutable: gpioExecutable,
	}
}

func (g gpio) Read(pin uint) (string, error) {
	args := []string{gpioReadCommand, tostr(pin)}
	return g.osHelper.Exec(g.gpioExecutable, args...)
}

func (g gpio) Write(pin uint, state string) error {
	return nil
}

func tostr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
