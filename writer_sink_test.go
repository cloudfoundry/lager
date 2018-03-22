package lager_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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

		BeforeEach(func() {
			wg := new(sync.WaitGroup)
			for i := 0; i < MaxThreads; i++ {
				wg.Add(1)
				go func() {
					sink.Log(lager.LogFormat{LogLevel: lager.INFO, Message: content})
					wg.Done()
				}()
			}
			wg.Wait()
		})

		It("writes to the given writer", func() {
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

var _ = Describe("PrettyPrintWriter", func() {
	const MaxThreads = 100

	var buf *bytes.Buffer
	var sink lager.Sink
	var message lager.PrettyFormat

	BeforeEach(func() {
		message = lager.PrettyFormat{}
		buf = new(bytes.Buffer)
		sink = lager.NewPrettySink(buf, lager.INFO)
	})

	It("renders in order: timestamp (in UTC), level, source, message and data fields", func() {
		expectedTime := time.Unix(0, 0)
		sink.Log(lager.LogFormat{
			LogLevel:  lager.INFO,
			Timestamp: formatTimestamp(expectedTime),
		})
		logBuf := gbytes.BufferWithBytes(buf.Bytes())
		Expect(logBuf).To(gbytes.Say(`{`))
		Expect(logBuf).To(gbytes.Say(`"timestamp":"1970-01-01T00:00:00Z",`))
		Expect(logBuf).To(gbytes.Say(`"level":"info",`))
		Expect(logBuf).To(gbytes.Say(`"source":"",`))
		Expect(logBuf).To(gbytes.Say(`"message":"",`))
		Expect(logBuf).To(gbytes.Say(`"data":null`))
		Expect(logBuf).To(gbytes.Say(`}`))
	})

	Context("when the internal time field of the provided log is zero", func() {
		testTimestamp := func(expected time.Time) {
			expected = expected.UTC()
			Expect(json.Unmarshal(buf.Bytes(), &message)).To(Succeed())
			Expect(time.Time(message.Timestamp)).To(BeTemporally("~", expected))
		}

		Context("and the unix epoch is set", func() {
			It("parses the timestamp", func() {
				expectedTime := time.Now().Add(time.Hour)
				sink.Log(lager.LogFormat{
					LogLevel:  lager.INFO,
					Timestamp: formatTimestamp(expectedTime),
				})
				testTimestamp(expectedTime)
			})
		})

		Context("the unix epoch is empty or invalid", func() {
			var invalidTimestamps = []string{
				"",
				"invalid",
				".123",
				"123.",
				"123.456.",
				"123.456.789",
				strconv.FormatInt(time.Now().Unix(), 10),         // invalid - missing "."
				strconv.FormatInt(-time.Now().Unix(), 10) + ".0", // negative
				time.Now().Format(time.RFC3339),
				time.Now().Format(time.RFC3339Nano),
			}

			It("uses the current time", func() {
				for _, ts := range invalidTimestamps {
					buf.Reset()
					sink.Log(lager.LogFormat{
						Timestamp: ts,
						LogLevel:  lager.INFO,
					})
					testTimestamp(time.Now())
				}
			})
		})
	})

	Context("when logging at or above the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.LogFormat{LogLevel: lager.INFO, Message: "hello world"})
		})

		It("writes to the given writer", func() {
			Expect(json.Unmarshal(buf.Bytes(), &message)).To(Succeed())
			Expect(message.Level).To(Equal("info"))
			Expect(message.Message).To(Equal("hello world"))
		})
	})

	Context("when a unserializable object is passed into data", func() {
		BeforeEach(func() {
			invalid := lager.LogFormat{
				LogLevel: lager.INFO,
				Message:  "hello world",
				Data:     lager.Data{"nope": func() {}},
			}
			sink.Log(invalid)
		})

		It("logs the serialization error", func() {
			json.Unmarshal(buf.Bytes(), &message)
			Expect(message.Message).To(Equal("hello world"))
			Expect(message.Level).To(Equal("info"))
			Expect(message.Data["lager serialisation error"]).To(Equal("json: unsupported type: func()"))
			Expect(message.Data["data_dump"]).ToNot(BeEmpty())
		})
	})

	Context("when logging below the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.LogFormat{LogLevel: lager.DEBUG, Message: "hello world"})
		})

		It("does not write to the given writer", func() {
			Expect(buf).To(Equal(bytes.NewBuffer(nil)))
		})
	})

	Context("when logging from multiple threads", func() {
		var content = "abcdefg "

		BeforeEach(func() {
			wg := new(sync.WaitGroup)
			for i := 0; i < MaxThreads; i++ {
				wg.Add(1)
				go func() {
					sink.Log(lager.LogFormat{LogLevel: lager.INFO, Message: content})
					wg.Done()
				}()
			}
			wg.Wait()
		})

		It("writes to the given writer", func() {
			lines := strings.Split(string(buf.Bytes()), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				var lineBuffer = bytes.NewBufferString(line)
				Expect(json.Unmarshal(lineBuffer.Bytes(), &message)).To(Succeed())
				Expect(message.Level).To(Equal("info"))
				Expect(message.Message).To(Equal(content))
			}
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

// duplicate of logger.go's formatTimestamp() function
func formatTimestamp(t time.Time) string {
	return fmt.Sprintf("%.9f", float64(t.UnixNano())/1e9)
}
