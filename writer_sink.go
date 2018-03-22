package lager

import (
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// A Sink represents a write destination for a Logger. It provides
// a thread-safe interface for writing logs
type Sink interface {
	//Log to the sink.  Best effort -- no need to worry about errors.
	Log(LogFormat)
}

type writerSink struct {
	writer      io.Writer
	minLogLevel LogLevel
	writeL      *sync.Mutex
}

func NewWriterSink(writer io.Writer, minLogLevel LogLevel) Sink {
	return &writerSink{
		writer:      writer,
		minLogLevel: minLogLevel,
		writeL:      new(sync.Mutex),
	}
}

func (sink *writerSink) Log(log LogFormat) {
	if log.LogLevel < sink.minLogLevel {
		return
	}

	sink.writeL.Lock()
	sink.writer.Write(log.ToJSON())
	sink.writer.Write([]byte("\n"))
	sink.writeL.Unlock()
}

type prettySink struct {
	writer      io.Writer
	minLogLevel LogLevel
	writeL      sync.Mutex
}

func NewPrettySink(writer io.Writer, minLogLevel LogLevel) Sink {
	return &prettySink{
		writer:      writer,
		minLogLevel: minLogLevel,
	}
}

func (sink *prettySink) Log(log LogFormat) {
	if log.LogLevel < sink.minLogLevel {
		return
	}
	t := log.time
	if t.IsZero() {
		t = parseTimestamp(log.Timestamp)
	}
	out := PrettyFormat{
		Timestamp: RFC3339Time(t),
		Level:     log.LogLevel.String(),
		Source:    log.Source,
		Message:   log.Message,
		Data:      log.Data,
		Error:     log.Error,
	}
	sink.writeL.Lock()
	sink.writer.Write(out.ToJSON())
	sink.writer.Write([]byte("\n"))
	sink.writeL.Unlock()
}

func parseTimestamp(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	n := strings.IndexByte(s, '.')
	if n <= 0 || n == len(s)-1 {
		return time.Now()
	}
	sec, err := strconv.ParseInt(s[:n], 10, 64)
	if err != nil || sec < 0 {
		return time.Now()
	}
	nsec, err := strconv.ParseInt(s[n+1:], 10, 64)
	if err != nil || nsec < 0 {
		return time.Now()
	}
	return time.Unix(sec, nsec)
}
