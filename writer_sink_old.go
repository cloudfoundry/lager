// +build !new

package lager

func (sink *writerSink) Log(log LogFormat) {
	if log.LogLevel < sink.minLogLevel {
		return
	}

	sink.writeL.Lock()
	sink.writer.Write(log.ToJSON())
	sink.writer.Write([]byte("\n"))
	sink.writeL.Unlock()
}
