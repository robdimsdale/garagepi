package oshelper

import (
	"time"

	"github.com/pivotal-golang/lager"
)

type OsHelper interface {
	Sleep(d time.Duration)
}

type OsHelperImpl struct {
	logger lager.Logger
}

func NewOsHelperImpl(
	logger lager.Logger,
) *OsHelperImpl {
	return &OsHelperImpl{
		logger: logger,
	}
}

func (h *OsHelperImpl) Sleep(d time.Duration) {
	h.logger.Info("sleeping", lager.Data{"duration": d.String()})
	time.Sleep(d)
}
