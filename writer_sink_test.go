package lager_test

import (
	"runtime"
	"sync"
	"time"

	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WriterSink", func() {
	const MaxThreads = 100

	var sink lager.Sink
	var writer *copyWriter

	BeforeSuite(func() {
		runtime.GOMAXPROCS(MaxThreads)
	})

	BeforeEach(func() {
		writer = NewCopyWriter()
		sink = lager.NewWriterSink(writer, lager.INFO)
	})

	JustBeforeEach(func() {
		// logs are written in a background thread with no confirmation,
		// so we must wait a bit to ensure any final writes are made
		time.Sleep(10 * time.Millisecond)
	})

	Context("when logging above the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.INFO, []byte("hello world"))
		})

		It("writes to the given writer", func() {
			Ω(writer.Copy()).Should(Equal([]byte("hello world\n")))
		})
	})

	Context("when logging below the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.DEBUG, []byte("hello world"))
		})

		It("does not write to the given writer", func() {
			Ω(writer.Copy()).Should(Equal([]byte{}))
		})
	})

	Context("when the writer attached to the sink is writing very slowly", func() {
		BeforeEach(func() {
			slowWriter := NewSlowWriter()
			sink = lager.NewWriterSink(slowWriter, lager.INFO)
		})

		It("should not block on calls to sink.Log", func(done Done) {
			for i := 0; i < 1024*2; i++ {
				sink.Log(lager.INFO, []byte("hello world"))
			}

			close(done)
		})

		It("waits for it when flushing", func() {
			sink.Log(lager.INFO, []byte("hello world"))

			t1 := time.Now()

			sink.Flush()

			Ω(time.Now().Sub(t1)).Should(BeNumerically(">", 500 * time.Millisecond))
		})
	})

	Context("when logging from multiple threads", func() {
		var content = "abcdefg "

		BeforeEach(func() {
			wg := new(sync.WaitGroup)
			for i := 0; i < MaxThreads; i++ {
				wg.Add(1)
				go func() {
					sink.Log(lager.INFO, []byte(content))
					wg.Done()
				}()
			}
			wg.Wait()
		})

		It("writes to the given writer", func() {
			expectedBytes := []byte{}
			for i := 0; i < MaxThreads; i++ {
				expectedBytes = append(expectedBytes, []byte(content)...)
				expectedBytes = append(expectedBytes, []byte("\n")...)
			}
			Ω(writer.Copy()).Should(Equal(expectedBytes))
		})
	})
})

// copyWriter is an INTENTIONALLY UNSAFE writer. Use it to test code that
// should be handling thread safety.
type copyWriter struct {
	contents []byte
	lock     *sync.RWMutex
}

func NewCopyWriter() *copyWriter {
	return &copyWriter{
		contents: []byte{},
		lock:     new(sync.RWMutex),
	}
}

// no, we really mean RLock on write.
func (writer *copyWriter) Write(p []byte) (n int, err error) {
	writer.lock.RLock()
	defer writer.lock.RUnlock()

	writer.contents = append(writer.contents, p...)
	return len(p), nil
}

func (writer *copyWriter) Copy() []byte {
	writer.lock.Lock()
	defer writer.lock.Unlock()

	contents := make([]byte, len(writer.contents))
	copy(contents, writer.contents)
	return contents
}

type slowWriter struct {
}

func NewSlowWriter() *slowWriter {
	return &slowWriter{}
}

func (writer *slowWriter) Write(p []byte) (n int, err error) {
	time.Sleep(1 * time.Second)
	return 0, nil
}
