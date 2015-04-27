package oshelper

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/robdimsdale/garagepi/logger"
)

type OsHelper interface {
	Exec(executable string, arg ...string) (string, error)
	Sleep(d time.Duration)
}

type OsHelperImpl struct {
	logger logger.Logger
}

func NewOsHelperImpl(
	logger logger.Logger,
) *OsHelperImpl {
	return &OsHelperImpl{
		logger: logger,
	}
}

func (h *OsHelperImpl) Exec(executable string, arg ...string) (string, error) {
	h.logger.Log(fmt.Sprintf("Executing: '%s %s'", executable, strings.Join(arg, " ")))
	out, err := exec.Command(executable, arg...).CombinedOutput()
	if err != nil {
		h.logger.Log(fmt.Sprintf("Error executing: '%s %s'", executable, strings.Join(arg, " ")))
	}
	return string(out), err
}

func (h *OsHelperImpl) Sleep(d time.Duration) {
	h.logger.Log("sleeping for " + d.String())
	time.Sleep(d)
}
