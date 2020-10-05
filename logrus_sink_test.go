package lager

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestNewLogrusSink(t *testing.T) {
	logger, _ := test.NewNullLogger()

	sink := NewLogrusSink(logger)

	assert.Equal(t, logger, sink.logger)
}

func TestLogrusSink_Log(t *testing.T) {
	const (
		source    = "test-source"
		message   = "test-message"
		key1      = "key1"
		value1    = "value1"
		key2      = "key2"
		value2    = "value2"
		timestamp = "some-timestamp"
	)

	data := Data{
		key1: value1,
		key2: value2,
	}

	err := errors.New("test-error")

	logger, hook := test.NewNullLogger()

	sink := NewLogrusSink(logger)

	sink.Log(LogFormat{
		Timestamp: timestamp,
		Source:    source,
		Message:   message,
		LogLevel:  INFO,
		Data:      data,
		Error:     err,
	})

	assert.Len(t, hook.Entries, 1)

	entry := hook.Entries[0]

	assert.Equal(t, timestamp, entry.Data["timestamp"])
	assert.Equal(t, source, entry.Data["source"])
	assert.Equal(t, message, entry.Message)
	assert.Equal(t, logrus.InfoLevel, entry.Level)
	assert.Equal(t, value1, entry.Data[key1])
	assert.Equal(t, value2, entry.Data[key2])
}

func Test_getLogrusLevel(t *testing.T) {
	cases := []struct {
		input       LogLevel
		expected    logrus.Level
		errExpected bool
	}{
		{
			input:    DEBUG,
			expected: logrus.DebugLevel,
		},
		{
			input:    INFO,
			expected: logrus.InfoLevel,
		},
		{
			input:    ERROR,
			expected: logrus.ErrorLevel,
		},
		{
			input:    FATAL,
			expected: logrus.FatalLevel,
		},
		{
			input:       -1,
			errExpected: true,
		},
	}

	for _, c := range cases {
		out, err := getLogrusLevel(c.input)

		if c.errExpected {
			assert.NotNil(t, err)
		}

		assert.Equal(t, c.expected, out)
	}
}
