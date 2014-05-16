package lager

import "io"

const logBufferSize = 1024

// A Sink represents a write destination for a Logger. It provides
// a thread-safe interface for writing logs
type Sink interface {
	//Log to the sink.  Best effort -- no need to worry about errors.
	Log(level LogLevel, payload []byte)
}

type writerSink struct {
	writer      io.Writer
	minLogLevel LogLevel
	logChan     chan []byte
}

func NewWriterSink(writer io.Writer, minLogLevel LogLevel) Sink {
	sink := &writerSink{
		writer:      writer,
		minLogLevel: minLogLevel,
		logChan:     make(chan []byte, logBufferSize),
	}

	go sink.listen()

	return sink
}

func (sink *writerSink) listen() {
	for {
		log := <-sink.logChan
		sink.writer.Write(log)
	}
}

func (sink *writerSink) Log(level LogLevel, log []byte) {
	if level < sink.minLogLevel {
		return
	}
	sink.logChan <- log
}
