package lager

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
)

type TestLogger struct {
	Logger
	*TestSink
}

func NewTestLogger(component string) *TestLogger {
	logger := NewLogger(component)
	testSink := NewTestSink()
	logger.RegisterSink(testSink)

	return &TestLogger{logger, testSink}
}

type TestSink struct {
	contents []byte
	lock     *sync.Mutex
	flushed  int
}

func NewTestSink() *TestSink {
	return &TestSink{
		lock: &sync.Mutex{},
	}
}

func (l *TestSink) Log(level LogLevel, p []byte) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.contents = append(l.contents, p...)
}

func (l *TestSink) Flush() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.flushed++
}

func (l *TestSink) Flushed() int {
	l.lock.Lock()
	defer l.lock.Unlock()

	return l.flushed
}

func (l *TestSink) Buffer() *bytes.Buffer {
	l.lock.Lock()
	defer l.lock.Unlock()

	contents := make([]byte, len(l.contents))
	copy(contents, l.contents)
	return bytes.NewBuffer(contents)
}

func (l *TestSink) Logs() []LogFormat {
	logs := []LogFormat{}
	decoder := json.NewDecoder(l.Buffer())
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
