package lager

import (
	"encoding/json"
)

type redactingSink struct {
	sink         Sink
	jsonRedacter *JSONRedacter
}

func NewRedactingSink(sink Sink, keyPatterns []string, valuePatterns []string) (Sink, error) {
	jsonRedacter, err := NewJSONRedacter(keyPatterns, valuePatterns)
	if err != nil {
		return nil, err
	}

	return &redactingSink{
		sink:         sink,
		jsonRedacter: jsonRedacter,
	}, nil
}

func (sink *redactingSink) Log(log LogFormat) {
	rawJSON, err := json.Marshal(log.Data)
	if err != nil {
		log.Data = dataForJSONMarhallingError(err, log.Data)

		rawJSON, err = json.Marshal(log.Data)
		if err != nil {
			panic(err)
		}
	}

	redactedJSON := sink.jsonRedacter.Redact(rawJSON)

	err = json.Unmarshal(redactedJSON, &log.Data)
	if err != nil {
		panic(err)
	}

	sink.sink.Log(log)
}
