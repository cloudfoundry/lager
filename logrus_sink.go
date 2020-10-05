package lager

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogrusSink struct {
	logger logrus.FieldLogger
}

func NewLogrusSink(logger logrus.FieldLogger) *LogrusSink {
	return &LogrusSink{logger: logger}
}

func (l *LogrusSink) Log(entry LogFormat) {
	level, err := getLogrusLevel(entry.LogLevel)
	if err != nil {
		l.logger.WithError(err).Errorf("Using the %s level", logrus.ErrorLevel.String())
		level = logrus.ErrorLevel
	}

	fields := logrus.Fields(entry.Data)
	fields["source"] = entry.Source
	fields["timestamp"] = entry.Timestamp

	logrusEntry := l.logger.WithFields(fields)

	if entry.Error != nil {
		logrusEntry = logrusEntry.WithError(entry.Error)
	}

	logrusEntry.Log(level, entry.Message)
}

func getLogrusLevel(lagerLevel LogLevel) (level logrus.Level, err error) {
	switch lagerLevel {
	case DEBUG:
		level = logrus.DebugLevel
	case INFO:
		level = logrus.InfoLevel
	case ERROR:
		level = logrus.ErrorLevel
	case FATAL:
		level = logrus.FatalLevel
	default:
		err = fmt.Errorf("%q: unhandled log level", lagerLevel.String())
	}

	return
}
