package oshelper

import (
	"time"

	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . OsHelper

type OsHelper interface {
	Sleep(d time.Duration)
}

type osHelperImpl struct {
	logger lager.Logger
}

func NewOsHelperImpl(
	logger lager.Logger,
) OsHelper {
	return &osHelperImpl{
		logger: logger,
	}
}

func (h *osHelperImpl) Sleep(d time.Duration) {
	h.logger.Info("sleeping", lager.Data{"duration": d.String()})
	time.Sleep(d)
}
