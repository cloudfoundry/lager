package lager

import "log/syslog"

type syslogPayload struct {
	level   LogLevel
	payload string
}

type syslogSink struct {
	minLogLevel LogLevel
	writer      *syslog.Writer
	logChan     chan syslogPayload
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
			sink.writer.Emerg(string(payload.payload))
		}
	}
}

func (sink *syslogSink) Log(level LogLevel, log []byte) {
	if level < sink.minLogLevel {
		return
	}
	payload := syslogPayload{
		level:   level,
		payload: string(log),
	}
	select {
	case sink.logChan <- payload:
	default:
	}
}
