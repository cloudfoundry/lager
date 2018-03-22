package lager

import (
	"encoding/json"
	"fmt"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
	FATAL
)

var logLevelStr = [...]string{
	DEBUG: "debug",
	INFO:  "info",
	ERROR: "error",
	FATAL: "fatal",
}

func (l LogLevel) String() string {
	if DEBUG <= l && l <= FATAL {
		return logLevelStr[l]
	}
	return "invalid"
}

type Data map[string]interface{}

type RFC3339Time time.Time

func (t RFC3339Time) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf(`"%s"`, time.Time(t).UTC().Format(time.RFC3339Nano))
	return []byte(stamp), nil
}

func (t *RFC3339Time) UnmarshalJSON(data []byte) error {
	return (*time.Time)(t).UnmarshalJSON(data)
}

type PrettyFormat struct {
	Timestamp RFC3339Time `json:"timestamp"`
	Level     string      `json:"level"`
	Source    string      `json:"source"`
	Message   string      `json:"message"`
	Data      Data        `json:"data"`
	Error     error       `json:"-"`
}

func (log PrettyFormat) ToJSON() []byte {
	content, err := json.Marshal(log)
	if err != nil {
		_, ok1 := err.(*json.UnsupportedTypeError)
		_, ok2 := err.(*json.MarshalerError)
		if ok1 || ok2 {
			log.Data = map[string]interface{}{
				"lager serialisation error": err.Error(),
				"data_dump":                 fmt.Sprintf("%#v", log.Data),
			}
			content, err = json.Marshal(log)
		}
		if err != nil {
			panic(err)
		}
	}
	return content
}

type LogFormat struct {
	Timestamp string   `json:"timestamp"`
	Source    string   `json:"source"`
	Message   string   `json:"message"`
	LogLevel  LogLevel `json:"log_level"`
	Data      Data     `json:"data"`
	Error     error    `json:"-"`
	time      time.Time
}

func (log LogFormat) ToJSON() []byte {
	content, err := json.Marshal(log)
	if err != nil {
		_, ok1 := err.(*json.UnsupportedTypeError)
		_, ok2 := err.(*json.MarshalerError)
		if ok1 || ok2 {
			log.Data = map[string]interface{}{"lager serialisation error": err.Error(), "data_dump": fmt.Sprintf("%#v", log.Data)}
			content, err = json.Marshal(log)
		}
		if err != nil {
			panic(err)
		}
	}
	return content
}
