// +build new

package lager

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func (sink *writerSink) Log(log LogFormat) {
	if log.LogLevel < sink.minLogLevel {
		return
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	buf.WriteRune('{')
	writeKV(buf, "timestamp", log.Timestamp)
	buf.WriteRune(',')
	writeKV(buf, "source", log.Source)
	buf.WriteRune(',')
	writeKV(buf, "message", log.Message)
	buf.WriteRune(',')
	writeLevel(buf, "log_level", log.LogLevel)
	buf.WriteRune(',')
	writeData(buf, log.Data)
	buf.WriteRune('}')
	buf.WriteString("\n")

	sink.writeL.Lock()
	sink.writer.Write(buf.Bytes())
	sink.writeL.Unlock()

	bufPool.Put(buf)
}

func writeLevel(w *bytes.Buffer, k string, v LogLevel) {
	writeString(w, k)
	w.WriteRune(':')
	w.WriteString(strconv.Itoa(int(v)))
}

func writeKV(w *bytes.Buffer, k string, v string) {
	writeString(w, k)
	w.WriteRune(':')
	writeString(w, v)
}

func writeString(w *bytes.Buffer, s string) {
	w.WriteRune('"')
	w.WriteString(s)
	w.WriteRune('"')
}

func writeValue(w *bytes.Buffer, v interface{}) {
	switch val := v.(type) {
	case string:
		r := strings.NewReplacer("\n", "\\n", "\t", "\\t", "\"", `\\"`)
		w.WriteRune('"')
		r.WriteString(w, val)
		w.WriteRune('"')

	case int:
		w.WriteString(strconv.Itoa(val))

	default:
		writeString(w, "UNKNOWN TYPE")
	}
}

func writeData(w *bytes.Buffer, d Data) {
	writeString(w, "data")
	w.WriteRune(':')
	size := len(d)

	if size == 0 {
		w.WriteString("null")
		return
	}

	w.WriteRune('{')

	i := 0

	for k, v := range d {
		writeString(w, k)
		w.WriteRune(':')
		writeValue(w, v)
		i++
		if i < size {
			w.WriteRune(',')
		}
	}

	w.WriteRune('}')
}
