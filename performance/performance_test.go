package performance

import (
	"io/ioutil"
	"os"
	"testing"

	"code.cloudfoundry.org/lager"
)

func BenchmarkLogWriting(b *testing.B) {
	temp := newTempFile(b)
	defer temp.cleanup()

	logger := lager.NewLogger("performance")
	logger.RegisterSink(lager.NewWriterSink(temp, lager.DEBUG))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Debug("thing", lager.Data{"example": "data"})
	}
}

type tempfile struct {
	file *os.File
	b    *testing.B
}

func newTempFile(b *testing.B) *tempfile {
	temp, err := ioutil.TempFile("", "lager_test")
	if err != nil {
		b.Error("failed to create log file:", err)
	}

	return &tempfile{
		file: temp,
		b:    b,
	}
}

func (t *tempfile) Write(bs []byte) (n int, err error) {
	return t.file.Write(bs)
}

func (t *tempfile) cleanup() {
	if err := t.file.Close(); err != nil {
		t.b.Error("failed to close log file", err)
	}

	if err := os.Remove(t.file.Name()); err != nil {
		t.b.Error("failed to close log file", err)
	}
}
