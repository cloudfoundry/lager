package lager

import (
	"io"
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

type writerSinkV2 struct {
	writer      io.Writer
	minLogLevel LogLevel
	writeL      *sync.Mutex
}

// NewWriterSinkV2 returns a sink that logs messages in the LogFormatV2
func NewWriterSinkV2(writer io.Writer, minLogLevel LogLevel) Sink {
	return &writerSinkV2{
		writer:      writer,
		minLogLevel: minLogLevel,
		writeL:      new(sync.Mutex),
	}
}

func (sink *writerSinkV2) Log(log LogFormat) {
	if log.LogLevel < sink.minLogLevel {
		return
	}
	t := log.time
	if t.IsZero() {
		// TODO: Parse the Timestamp field.  This check is required because
		// the time field is not exported - exporting it would change the
		// signature of the LogFormat struct.
		t = time.Now()
	}
	out := LogFormatV2{
		Timestamp: log.time,
		Source:    log.Source,
		Message:   log.Message,
		Level:     log.LogLevel.String(),
		Data:      log.Data,
		Error:     log.Error,
	}
	sink.writeL.Lock()
	sink.writer.Write(out.ToJSON())
	sink.writer.Write([]byte("\n"))
	sink.writeL.Unlock()
}
