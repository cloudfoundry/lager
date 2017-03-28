package lager

import (
	"encoding/json"
	"regexp"
)


const awsAccessKeyIDPattern = `AKIA[A-Z0-9]{16}`
const awsSecretAccessKeyPattern = `KEY["']?\s*(?::|=>|=)\s*["']?[A-Z0-9/\+=]{40}["']?`
const cryptMD5Pattern = `\$1\$[A-Z0-9./]{1,16}\$[A-Z0-9./]{22}`
const cryptSHA256Pattern = `\$5\$[A-Z0-9./]{1,16}\$[A-Z0-9./]{43}`
const cryptSHA512Pattern = `\$6\$[A-Z0-9./]{1,16}\$[A-Z0-9./]{86}`
const privateKeyHeaderPattern = `-----BEGIN(.*)PRIVATE KEY-----`

type JsonRedacter struct {
	keyMatchers []*regexp.Regexp
	valueMatchers []*regexp.Regexp
}

func NewJsonRedacter(keyPatterns []string, valuePatterns []string) (*JsonRedacter, error) {
	if keyPatterns == nil {
		keyPatterns = []string{"[Pp]wd","[Pp]ass"}
	}
	if valuePatterns == nil {
		valuePatterns = []string{awsAccessKeyIDPattern, awsSecretAccessKeyPattern, cryptMD5Pattern, cryptSHA256Pattern, cryptSHA512Pattern, privateKeyHeaderPattern}
	}
	ret := &JsonRedacter{}
	for _ ,v := range keyPatterns {
		r, err := regexp.Compile(v)
		if err != nil {
			return nil, err
		}
		ret.keyMatchers = append(ret.keyMatchers, r)
	}
	for _ ,v := range valuePatterns {
		r, err := regexp.Compile(v)
		if err != nil {
			return nil, err
		}
		ret.valueMatchers = append(ret.valueMatchers, r)
	}
	return ret, nil
}

func (r JsonRedacter) Redact(data []byte) ([]byte, error) {
	var jsonBlob interface{}
	err := json.Unmarshal(data, &jsonBlob)
	if err != nil {
		return nil, err
	}
	r.redactValue(&jsonBlob)

	data, err = json.Marshal(jsonBlob)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r JsonRedacter) redactValue(data *interface{}) interface{} {
	if data == nil {
		return data
	}

	if a, ok := (*data).([]interface{}); ok {
		r.redactArray(&a)
	} else if m, ok := (*data).(map[string]interface{}); ok {
		r.redactObject(&m)
	} else if s, ok := (*data).(string); ok {
		for _, m := range r.valueMatchers {
			if m.MatchString(s) {
				(*data) = "*REDACTED*"
				break
			}
		}
	}
	return (*data)
}

func (r JsonRedacter) redactArray(data *[]interface{}) {
	for i, _ := range *data {
		r.redactValue(&((*data)[i]))
	}
}

func (r JsonRedacter) redactObject(data *map[string]interface{}) {
	for k, v := range *data {
		for _, m := range r.keyMatchers {
			if m.MatchString(k) {
				(*data)[k] = "*REDACTED*"
				break
			}
		}
		if (*data)[k] != "*REDACTED*" {
			(*data)[k] = r.redactValue(&v)
		}
	}
}
