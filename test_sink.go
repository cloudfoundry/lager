package lager

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/onsi/gomega/gbytes"
)

type TestLogger struct {
	Logger
	*TestSink
}

type TestSink struct {
	Sink
	*gbytes.Buffer
}

func NewTestLogger(component string) *TestLogger {
	logger := NewLogger(component)

	testSink := NewTestSink()

	logger.RegisterSink(testSink)

	return &TestLogger{logger, testSink}
}

func NewTestSink() *TestSink {
	buffer := gbytes.NewBuffer()

	return &TestSink{
		Sink:   NewWriterSink(buffer, DEBUG),
		Buffer: buffer,
	}
}

func (s *TestSink) Logs() []LogFormat {
	logs := []LogFormat{}

	decoder := json.NewDecoder(bytes.NewBuffer(s.Buffer.Contents()))
	for {
		var log LogFormat
		if err := decoder.Decode(&log); err == io.EOF {
			return logs
		} else if err != nil {
			panic(err)
		}
		logs = append(logs, log)
	}

	return logs
}
