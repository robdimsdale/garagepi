package loglevel

import (
	"io/ioutil"
	"net/http"

	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/garagepi/logger"
)

func (s *Server) SetMinLevel(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var level lager.LogLevel

	switch logger.LogLevel(string(body)) {
	case logger.LogLevelDebug:
		level = lager.DEBUG
	case logger.LogLevelInfo:
		level = lager.INFO
	case logger.LogLevelError:
		level = lager.ERROR
	case logger.LogLevelFatal:
		level = lager.FATAL
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.sink.SetMinLevel(level)
}
