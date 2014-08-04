package chug

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pivotal-golang/lager"
)

type Entry struct {
	IsLager bool
	Raw     []byte
	Log     LogEntry
}

func Chug(reader io.Reader, out chan<- Entry) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		out <- entry(scanner.Bytes())
	}
}

func entry(raw []byte) (entry Entry) {
	entry = Entry{
		IsLager: false,
		Raw:     raw,
	}

	rawString := string(raw)
	idx := strings.Index(rawString, "{")
	if idx == -1 {
		return
	}

	var lagerLog lager.LogFormat
	decoder := json.NewDecoder(strings.NewReader(rawString[idx:]))
	err := decoder.Decode(&lagerLog)
	if err != nil {
		return
	}

	entry.Log, entry.IsLager = convertLagerLog(lagerLog)

	return
}

func convertLagerLog(lagerLog lager.LogFormat) (LogEntry, bool) {
	timestamp, err := strconv.ParseFloat(lagerLog.Timestamp, 64)

	if err != nil {
		return LogEntry{}, false
	}

	data := lagerLog.Data

	var logErr error
	dataErr, ok := lagerLog.Data["error"]
	if ok {
		errorString, ok := dataErr.(string)
		if !ok {
			return LogEntry{}, false
		}
		logErr = errors.New(errorString)
		delete(lagerLog.Data, "error")
	}

	var logTrace string
	dataTrace, ok := lagerLog.Data["trace"]
	if ok {
		logTrace, ok = dataTrace.(string)
		if !ok {
			return LogEntry{}, false
		}
		delete(lagerLog.Data, "trace")
	}

	var logJoinedSession string
	dataSession, ok := lagerLog.Data["session"]
	if ok {
		logJoinedSession, ok = dataSession.(string)
		if !ok {
			return LogEntry{}, false
		}
		delete(lagerLog.Data, "session")
	}

	messageComponents := strings.Split(lagerLog.Message, ".")
	sessionComponents := strings.Split(logJoinedSession, ".")

	var logAction string
	var logTasks []Task

	n := len(messageComponents)
	switch {
	case n <= 1:
		return LogEntry{}, false
	case n == 2:
		if logJoinedSession != "" {
			return LogEntry{}, false
		}
		logAction = messageComponents[len(messageComponents)-1]
	default:
		if len(messageComponents)-2 != len(sessionComponents) {
			return LogEntry{}, false
		}
		messageComponents = messageComponents[1:]
		logAction = messageComponents[len(messageComponents)-1]
		for i, session := range sessionComponents {
			logTasks = append(logTasks, Task{messageComponents[i], session})
		}
	}

	return LogEntry{
		Timestamp: time.Unix(0, int64(timestamp*1e9)),
		LogLevel:  lagerLog.LogLevel,
		Source:    lagerLog.Source,

		Action: logAction,
		Tasks:  logTasks,

		Error: logErr,
		Trace: logTrace,

		Data: data,
	}, true
}
