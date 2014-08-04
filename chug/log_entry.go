package chug

import (
	"strings"
	"time"

	"github.com/pivotal-golang/lager"
)

type Task struct {
	Task    string
	Session string
}

type LogEntry struct {
	Timestamp time.Time
	LogLevel  lager.LogLevel

	Source string

	Tasks  []Task
	Action string

	Error error
	Trace string

	Data lager.Data
}

func (l LogEntry) Message() string {
	message := []string{}
	for _, task := range l.Tasks {
		message = append(message, task.Task)
	}
	message = append(message, l.Action)

	return strings.Join(message, ".")
}

func (l LogEntry) Session() string {
	session := []string{}
	for _, task := range l.Tasks {
		session = append(session, task.Session)
	}
	return strings.Join(session, ".")
}
