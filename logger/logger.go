package logger

import (
	"log"
)

type Logger interface {
	Log(message string)
}

type LoggerImpl struct {
	loggingOn bool
}

func NewLoggerImpl(loggingOn bool) *LoggerImpl {
	return &LoggerImpl{
		loggingOn: loggingOn,
	}
}

func (l LoggerImpl) Log(message string) {
	if l.loggingOn {
		log.Println(message)
	}
}
