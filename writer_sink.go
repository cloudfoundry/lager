package lager

import (
	"io"
	"sync"
)

const logBufferSize = 1024

// A Sink represents a write destination for a Logger. It provides
// a thread-safe interface for writing logs
type Sink interface {
	//Log to the sink.  Best effort -- no need to worry about errors.
	Log(level LogLevel, payload []byte)

	//Wait for in-flight logs to be written.
	Flush()
}

type writerSink struct {
	writer      io.Writer
	minLogLevel LogLevel
	logChan     chan []byte
	flush       *sync.WaitGroup
}

func NewWriterSink(writer io.Writer, minLogLevel LogLevel) Sink {
	sink := &writerSink{
		writer:      writer,
		minLogLevel: minLogLevel,
		logChan:     make(chan []byte, logBufferSize),
		flush:       new(sync.WaitGroup),
	}

	go sink.listen()

	return sink
}

func (sink *writerSink) listen() {
	for {
		log := <-sink.logChan
		sink.writer.Write(log)
		sink.writer.Write([]byte("\n"))
		sink.flush.Done()
	}
}

func (sink *writerSink) Log(level LogLevel, log []byte) {
	if level < sink.minLogLevel {
		return
	}
	sink.flush.Add(1)

	select {
	case sink.logChan <- log:
	default:
		sink.flush.Done()
	}
}

func (sink *writerSink) Flush() {
	sink.flush.Wait()
}
