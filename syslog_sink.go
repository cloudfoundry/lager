package lager

import "log/syslog"

type syslogSink struct {
	minLogLevel LogLevel
	writer      *syslog.Writer
}

func NewSyslogSink(transport, serverAddress, tag string, minLogLevel LogLevel) (Sink, error) {
	writer, err := syslog.Dial(transport, serverAddress, syslog.LOG_EMERG|syslog.LOG_DAEMON, tag)
	if err != nil {
		return nil, err
	}
	return &syslogSink{
		minLogLevel: minLogLevel,
		writer:      writer,
	}, nil
}

func (sink *syslogSink) Log(level LogLevel, payload []byte) {
	if level < sink.minLogLevel {
		return
	}

	switch level {
	case DEBUG:
		sink.writer.Debug(string(payload))
	case INFO:
		sink.writer.Info(string(payload))
	case ERROR:
		sink.writer.Err(string(payload))
	case FATAL:
		sink.writer.Emerg(string(payload))
	}
}
