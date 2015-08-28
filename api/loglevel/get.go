package loglevel

import (
	"fmt"
	"net/http"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/logger"
)

func (s *Server) GetMinLevel(w http.ResponseWriter, r *http.Request) {
	var level logger.LogLevel

	switch s.sink.GetMinLevel() {
	case lager.DEBUG:
		level = logger.LogLevelDebug
	case lager.INFO:
		level = logger.LogLevelInfo
	case lager.ERROR:
		level = logger.LogLevelError
	case lager.FATAL:
		level = logger.LogLevelFatal
	default:
		s.logger.Error("unknown-log-level", nil, lager.Data{
			"level": level,
		})
		level = logger.LogLevelInvalid
	}

	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "%s", level)
}
