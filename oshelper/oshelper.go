package oshelper

import (
	"time"

	"github.com/robdimsdale/garagepi/logger"
)

type OsHelper interface {
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

func (h *OsHelperImpl) Sleep(d time.Duration) {
	h.logger.Log("sleeping for " + d.String())
	time.Sleep(d)
}
