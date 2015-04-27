package oshelper

import (
	"os/exec"
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
	out, err := exec.Command(executable, arg...).CombinedOutput()
	return string(out), err
}

func (h *OsHelperImpl) Sleep(d time.Duration) {
	h.logger.Log("sleeping for " + d.String())
	time.Sleep(d)
}
