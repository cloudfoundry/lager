package lager

import (
	"fmt"
	"time"
)

type Logger interface {
	RegisterSink(Sink)
	Debug(task, action, description string, data ...Data)
}

type logger struct {
	component string
	sinks     []Sink
}

func NewLogger(component string) Logger {
	return &logger{
		component: component,
		sinks:     []Sink{},
	}
}

func (l *logger) RegisterSink(sink Sink) {
	l.sinks = append(l.sinks, sink)
}

func (l *logger) Debug(task, action, description string, data ...Data) {
	var logData Data
	if len(data) > 0 {
		logData = data[0]
	}

	logData["description"] = description

	log := LogFormat{
		Timestamp: fmt.Sprintf("%.9f", time.Now().UnixNano()),
		Source:    l.component,
		Message:   fmt.Sprintf("%s.%s.%s", l.component, task, action),
		LogLevel:  DEBUG,
		Data:      logData,
	}

	for _, sink := range l.sinks {
		sink.Log(DEBUG, log.ToJSON())
	}
}
