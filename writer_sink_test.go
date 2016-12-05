package lager_test

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"

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

	Context("when logging above the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.LogFormat{LogLevel: lager.INFO, Message: "hello world"})
		})

		It("writes to the given writer", func() {
			Expect(writer.Copy()).To(MatchJSON(`{"message":"hello world","log_level":1,"timestamp":"","source":"","data":null}`))
		})
	})

	Context("when a unserializable object is passed into data", func() {
		BeforeEach(func() {
			sink.Log(lager.LogFormat{LogLevel: lager.INFO, Message: "hello world", Data: map[string]interface{}{"some_key": func() {}}})
		})

		It("logs the serialization error", func() {
			message := map[string]interface{}{}
			json.Unmarshal(writer.Copy(), &message)
			Expect(message["message"]).To(Equal("hello world"))
			Expect(message["log_level"]).To(Equal(float64(1)))
			Expect(message["data"].(map[string]interface{})["lager serialisation error"]).To(Equal("json: unsupported type: func()"))
			Expect(message["data"].(map[string]interface{})["data_dump"]).ToNot(BeEmpty())
		})

		Measure("should be efficient", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				for i := 0; i < 5000; i++ {
					sink.Log(lager.LogFormat{LogLevel: lager.INFO, Message: "hello world", Data: map[string]interface{}{"some_key": func() {}}})
					Expect(writer.Copy()).ToNot(BeEmpty())
				}
			})

			Expect(runtime.Seconds()).To(BeNumerically("<", 1), "logging shouldn't take too long.")
		}, 1)
	})

	Context("when logging below the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.LogFormat{LogLevel: lager.DEBUG, Message: "hello world"})
		})

		It("does not write to the given writer", func() {
			Expect(writer.Copy()).To(Equal([]byte{}))
		})
	})

	Context("when logging from multiple threads", func() {
		var content = "abcdefg "
		var wg *sync.WaitGroup

		JustBeforeEach(func() {
			wg = new(sync.WaitGroup)
			for i := 0; i < MaxThreads; i++ {
				wg.Add(1)
				go func() {
					sink.Log(lager.LogFormat{LogLevel: lager.INFO, Message: content})
					wg.Done()
				}()
			}
		})

		It("writes to the given writer", func() {
			wg.Wait()
			lines := strings.Split(string(writer.Copy()), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				Expect(line).To(MatchJSON(fmt.Sprintf(`{"message":"%s","log_level":1,"timestamp":"","source":"","data":null}`, content)))
			}
		})

		FContext("when the underlying writer is slow", func() {
			var slowWriter *slowWriter

			BeforeEach(func() {
				slowWriter = NewSlowWriter()
				sink = lager.NewWriterSink(slowWriter, lager.INFO)
			})

			It("does not block other log threads", func() {
				// WIP Note: this isn't working to generate a red-test yet. Needs more work.
				Eventually(func() bool { wg.Wait(); return true }, 1*time.Second).Should(BeTrue())
				slowWriter.channel <- "release"

				lines := strings.Split(string(writer.Copy()), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					Expect(line).To(MatchJSON(fmt.Sprintf(`{"message":"%s","log_level":1,"timestamp":"","source":"","data":null}`, content)))
				}
			})
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

// slowWriter

func NewSlowWriter() *slowWriter {
	return &slowWriter{
		channel: make(chan string),
		cpw:     NewCopyWriter(),
	}
}

type slowWriter struct {
	channel chan string
	cpw     *copyWriter
}

func (writer *slowWriter) Write(p []byte) (n int, err error) {
	<-writer.channel
	return writer.cpw.Write(p)
}

func (writer *slowWriter) Copy() []byte {
	return writer.cpw.Copy()
}
