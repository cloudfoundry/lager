//go:build go1.23

package lager

import (
	"context"
	"log/slog"
)

// Type slogSink wraps an slog.Logger as a Sink
type slogSink struct {
	logger *slog.Logger
}

// NewSlogSink wraps a slog.Logger as a lager Sink
// This allows code using slog to integrate with code that uses lager
// Note the following log level conversions:
//
//	lager.DEBUG -> slog.LevelDebug
//	lager.ERROR -> slog.LevelError
//	lager.FATAL -> slog.LevelError
//	default     -> slog.LevelInfo
func NewSlogSink(l *slog.Logger) Sink {
	return &slogSink{logger: l}
}

// Log exists to implement the lager.Sink interface.
func (l *slogSink) Log(f LogFormat) {
	// For lager.Error() and lager.Fatal() the error (and stacktrace) are already in f.Data
	r := slog.NewRecord(f.time, toSlogLevel(f.LogLevel), f.Message, 0)
	r.AddAttrs(toAttr(f.Data)...)

	// By calling the handler directly we can pass through the original timestamp,
	// whereas calling a method on the logger would generate a new timestamp
	l.logger.Handler().Handle(context.Background(), r)
}

// toAttr converts a lager.Data into []slog.Attr
func toAttr(d Data) []slog.Attr {
	l := len(d)
	if l == 0 {
		return nil
	}

	attr := make([]slog.Attr, 0, l)
	for k, v := range d {
		attr = append(attr, slog.Any(k, v))
	}

	return attr
}

// toSlogLevel converts lager log levels to slog levels
func toSlogLevel(l LogLevel) slog.Level {
	switch l {
	case DEBUG:
		return slog.LevelDebug
	case ERROR, FATAL:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// toSlogLevel converts slog levels to lager log levels
func toLagerLevel(l slog.Level) LogLevel {
	switch l {
	case slog.LevelDebug:
		return DEBUG
	case slog.LevelError:
		return ERROR
	case slog.LevelWarn:
		return ERROR
	case slog.LevelInfo:
		return INFO
	default:
		return INFO
	}
}

func (sink *slogSink) SetMinLevel(level LogLevel) {
	slog.SetLogLoggerLevel(toSlogLevel(level))
}

func (sink *slogSink) GetMinLevel() LogLevel {

	noopContext := context.TODO()

	// We have to use a drop through logic to find out the minium logging level of slog
	// slog's logger has no way of returning the current log level hence this logic.
	if sink.logger.Handler().Enabled(noopContext, slog.LevelDebug) {
		return DEBUG
	}

	if sink.logger.Handler().Enabled(noopContext, slog.LevelInfo) {
		return INFO
	}

	if sink.logger.Handler().Enabled(noopContext, slog.LevelWarn) {
		return ERROR
	}

	if sink.logger.Handler().Enabled(noopContext, slog.LevelError) {
		return ERROR
	}

	return INFO
}
