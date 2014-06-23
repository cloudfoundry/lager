package lager

import (
	"log/syslog"
	"sync"
)

type syslogPayload struct {
	level   LogLevel
	payload string
}

type syslogSink struct {
	minLogLevel LogLevel
	writer      *syslog.Writer
	logChan     chan syslogPayload
	flush       *sync.WaitGroup
}

func NewSyslogSink(transport, serverAddress, tag string, minLogLevel LogLevel) (Sink, error) {
	writer, err := syslog.Dial(transport, serverAddress, syslog.LOG_EMERG|syslog.LOG_DAEMON, tag)
	if err != nil {
		return nil, err
	}

	sink := &syslogSink{
		minLogLevel: minLogLevel,
		writer:      writer,
		logChan:     make(chan syslogPayload, logBufferSize), //no buffer as the underlying sylog pipe is buffered
		flush:       new(sync.WaitGroup),
	}

	go sink.listen()

	return sink, nil
}

func (sink *syslogSink) listen() {
	for {
		payload := <-sink.logChan

		switch payload.level {
		case DEBUG:
			sink.writer.Debug(string(payload.payload))
		case INFO:
			sink.writer.Info(string(payload.payload))
		case ERROR:
			sink.writer.Err(string(payload.payload))
		case FATAL:
			sink.writer.Crit(string(payload.payload))
		}

		sink.flush.Done()
	}
}

func (sink *syslogSink) Flush() {
	sink.flush.Wait()
}

func (sink *syslogSink) Log(level LogLevel, log []byte) {
	if level < sink.minLogLevel {
		return
	}

	payload := syslogPayload{
		level:   level,
		payload: string(log),
	}

	sink.flush.Add(1)

	select {
	case sink.logChan <- payload:
	default:
		sink.flush.Done()
	}
}
