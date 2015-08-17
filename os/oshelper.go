package os

import (
	"time"

	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter . OSHelper

type OSHelper interface {
	Sleep(d time.Duration)
}

type osHelper struct {
	logger lager.Logger
}

func NewOSHelper(
	logger lager.Logger,
) OSHelper {
	return &osHelper{
		logger: logger,
	}
}

func (h *osHelper) Sleep(d time.Duration) {
	h.logger.Info("sleeping", lager.Data{"duration": d.String()})
	time.Sleep(d)
}
