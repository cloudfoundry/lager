package lager

import (
	"io"
	"sync"
	"encoding/json"
)

type redactingWriterSink struct {
	writer      io.Writer
	minLogLevel LogLevel
	writeL      *sync.Mutex
}

func NewRedactingWriterSink(writer io.Writer, minLogLevel LogLevel) Sink {
	return &writerSink{
		writer:      writer,
		minLogLevel: minLogLevel,
		writeL:      new(sync.Mutex),
	}
}

func (sink *redactingWriterSink) Log(log LogFormat) {
	if log.LogLevel < sink.minLogLevel {
		return
	}

	sink.writeL.Lock()

  // first, create a generic object representation of the log data
	v := log.ToJSON()
	m := &map[string]interface{}{}
	json.Unmarshal(v, m)

	// then redact the data before logging it

	sink.writer.Write(log.ToJSON())
	sink.writer.Write([]byte("\n"))
	sink.writeL.Unlock()
}



