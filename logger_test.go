package lager_test

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var logger lager.Logger
	var testSink *lagertest.TestSink

	var component = "my-component"
	var action = "my-action"
	var logData = lager.Data{
		"foo":      "bar",
		"a-number": 7,
	}
	var anotherLogData = lager.Data{
		"baz":      "quux",
		"b-number": 43,
	}

	BeforeEach(func() {
		logger = lager.NewLogger(component)
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)
	})

	var TestCommonLogFeatures = func(level lager.LogLevel) {
		var log lager.LogFormat

		BeforeEach(func() {
			log = testSink.Logs()[0]
		})

		It("writes a log to the sink", func() {
			Expect(testSink.Logs()).To(HaveLen(1))
		})

		It("records the source component", func() {
			Expect(log.Source).To(Equal(component))
		})

		It("outputs a properly-formatted message", func() {
			Expect(log.Message).To(Equal(fmt.Sprintf("%s.%s", component, action)))
		})

		It("has a timestamp", func() {
			expectedTime := float64(time.Now().UnixNano()) / 1e9
			parsedTimestamp, err := strconv.ParseFloat(log.Timestamp, 64)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedTimestamp).To(BeNumerically("~", expectedTime, 1.0))
		})

		It("sets the proper output level", func() {
			Expect(log.LogLevel).To(Equal(level))
		})
	}

	var TestLogData = func() {
		var log lager.LogFormat

		BeforeEach(func() {
			log = testSink.Logs()[0]
		})

		It("data contains custom user data", func() {
			Expect(log.Data["foo"]).To(Equal("bar"))
			Expect(log.Data["a-number"]).To(BeNumerically("==", 7))
			Expect(log.Data["baz"]).To(Equal("quux"))
			Expect(log.Data["b-number"]).To(BeNumerically("==", 43))
		})
	}

	Describe("Session", func() {
		var session lager.Logger

		BeforeEach(func() {
			session = logger.Session("sub-action")
		})

		Describe("the returned logger", func() {
			JustBeforeEach(func() {
				session.Debug("some-debug-action", lager.Data{"level": "debug"})
				session.Info("some-info-action", lager.Data{"level": "info"})
				session.Error("some-error-action", errors.New("oh no!"), lager.Data{"level": "error"})

				defer func() {
					recover() //nolint:errcheck
				}()

				session.Fatal("some-fatal-action", errors.New("oh no!"), lager.Data{"level": "fatal"})
			})

			It("logs with a shared session id in the data", func() {
				Expect(testSink.Logs()[0].Data["session"]).To(Equal("1"))
				Expect(testSink.Logs()[1].Data["session"]).To(Equal("1"))
				Expect(testSink.Logs()[2].Data["session"]).To(Equal("1"))
				Expect(testSink.Logs()[3].Data["session"]).To(Equal("1"))
			})

			It("logs with the task added to the message", func() {
				Expect(testSink.Logs()[0].Message).To(Equal("my-component.sub-action.some-debug-action"))
				Expect(testSink.Logs()[1].Message).To(Equal("my-component.sub-action.some-info-action"))
				Expect(testSink.Logs()[2].Message).To(Equal("my-component.sub-action.some-error-action"))
				Expect(testSink.Logs()[3].Message).To(Equal("my-component.sub-action.some-fatal-action"))
			})

			It("logs with the original data", func() {
				Expect(testSink.Logs()[0].Data["level"]).To(Equal("debug"))
				Expect(testSink.Logs()[1].Data["level"]).To(Equal("info"))
				Expect(testSink.Logs()[2].Data["level"]).To(Equal("error"))
				Expect(testSink.Logs()[3].Data["level"]).To(Equal("fatal"))
			})

			Context("with data", func() {
				BeforeEach(func() {
					session = logger.Session("sub-action", lager.Data{"foo": "bar"})
				})

				It("logs with the data added to the message", func() {
					Expect(testSink.Logs()[0].Data["foo"]).To(Equal("bar"))
					Expect(testSink.Logs()[1].Data["foo"]).To(Equal("bar"))
					Expect(testSink.Logs()[2].Data["foo"]).To(Equal("bar"))
					Expect(testSink.Logs()[3].Data["foo"]).To(Equal("bar"))
				})

				It("keeps the original data", func() {
					Expect(testSink.Logs()[0].Data["level"]).To(Equal("debug"))
					Expect(testSink.Logs()[1].Data["level"]).To(Equal("info"))
					Expect(testSink.Logs()[2].Data["level"]).To(Equal("error"))
					Expect(testSink.Logs()[3].Data["level"]).To(Equal("fatal"))
				})
			})

			Context("with another session", func() {
				BeforeEach(func() {
					session = logger.Session("next-sub-action")
				})

				It("logs with a shared session id in the data", func() {
					Expect(testSink.Logs()[0].Data["session"]).To(Equal("2"))
					Expect(testSink.Logs()[1].Data["session"]).To(Equal("2"))
					Expect(testSink.Logs()[2].Data["session"]).To(Equal("2"))
					Expect(testSink.Logs()[3].Data["session"]).To(Equal("2"))
				})

				It("logs with the task added to the message", func() {
					Expect(testSink.Logs()[0].Message).To(Equal("my-component.next-sub-action.some-debug-action"))
					Expect(testSink.Logs()[1].Message).To(Equal("my-component.next-sub-action.some-info-action"))
					Expect(testSink.Logs()[2].Message).To(Equal("my-component.next-sub-action.some-error-action"))
					Expect(testSink.Logs()[3].Message).To(Equal("my-component.next-sub-action.some-fatal-action"))
				})
			})

			Describe("WithData", func() {
				BeforeEach(func() {
					session = logger.WithData(lager.Data{"foo": "bar"})
				})

				It("returns a new logger with the given data", func() {
					Expect(testSink.Logs()[0].Data["foo"]).To(Equal("bar"))
					Expect(testSink.Logs()[1].Data["foo"]).To(Equal("bar"))
					Expect(testSink.Logs()[2].Data["foo"]).To(Equal("bar"))
					Expect(testSink.Logs()[3].Data["foo"]).To(Equal("bar"))
				})

				It("does not append to the logger's task", func() {
					Expect(testSink.Logs()[0].Message).To(Equal("my-component.some-debug-action"))
				})
			})

			Context("with a nested session", func() {
				BeforeEach(func() {
					session = session.Session("sub-sub-action")
				})

				It("logs with a shared session id in the data", func() {
					Expect(testSink.Logs()[0].Data["session"]).To(Equal("1.1"))
					Expect(testSink.Logs()[1].Data["session"]).To(Equal("1.1"))
					Expect(testSink.Logs()[2].Data["session"]).To(Equal("1.1"))
					Expect(testSink.Logs()[3].Data["session"]).To(Equal("1.1"))
				})

				It("logs with the task added to the message", func() {
					Expect(testSink.Logs()[0].Message).To(Equal("my-component.sub-action.sub-sub-action.some-debug-action"))
					Expect(testSink.Logs()[1].Message).To(Equal("my-component.sub-action.sub-sub-action.some-info-action"))
					Expect(testSink.Logs()[2].Message).To(Equal("my-component.sub-action.sub-sub-action.some-error-action"))
					Expect(testSink.Logs()[3].Message).To(Equal("my-component.sub-action.sub-sub-action.some-fatal-action"))
				})
			})
		})
	})

	Describe("Debug", func() {
		Context("with log data", func() {
			BeforeEach(func() {
				logger.Debug(action, logData, anotherLogData)
			})

			TestCommonLogFeatures(lager.DEBUG)
			TestLogData()
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				logger.Debug(action)
			})

			TestCommonLogFeatures(lager.DEBUG)
		})
	})

	Describe("Info", func() {
		Context("with log data", func() {
			BeforeEach(func() {
				logger.Info(action, logData, anotherLogData)
			})

			TestCommonLogFeatures(lager.INFO)
			TestLogData()
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				logger.Info(action)
			})

			TestCommonLogFeatures(lager.INFO)
		})
	})

	Describe("Error", func() {
		var err = errors.New("oh noes!")
		Context("with log data", func() {
			BeforeEach(func() {
				logger.Error(action, err, logData, anotherLogData)
			})

			TestCommonLogFeatures(lager.ERROR)
			TestLogData()

			It("data contains error message", func() {
				Expect(testSink.Logs()[0].Data["error"]).To(Equal(err.Error()))
			})

			It("retains the original error values", func() {
				Expect(testSink.Errors).To(Equal([]error{err}))
			})
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				logger.Error(action, err)
			})

			TestCommonLogFeatures(lager.ERROR)

			It("data contains error message", func() {
				Expect(testSink.Logs()[0].Data["error"]).To(Equal(err.Error()))
			})

			It("retains the original error values", func() {
				Expect(testSink.Errors).To(Equal([]error{err}))
			})
		})

		Context("with no error", func() {
			BeforeEach(func() {
				logger.Error(action, nil)
			})

			TestCommonLogFeatures(lager.ERROR)

			It("does not contain the error message", func() {
				Expect(testSink.Logs()[0].Data).NotTo(HaveKey("error"))
			})
		})
	})

	Describe("Fatal", func() {
		var err = errors.New("oh noes!")
		var fatalErr interface{}

		Context("with log data", func() {
			BeforeEach(func() {
				defer func() {
					fatalErr = recover()
				}()

				logger.Fatal(action, err, logData, anotherLogData)
			})

			TestCommonLogFeatures(lager.FATAL)
			TestLogData()

			It("data contains error message", func() {
				Expect(testSink.Logs()[0].Data["error"]).To(Equal(err.Error()))
			})

			It("data contains stack trace", func() {
				Expect(testSink.Logs()[0].Data["trace"]).NotTo(BeEmpty())
			})

			It("panics with the provided error", func() {
				Expect(fatalErr).To(Equal(err))
			})

			It("retains the original error values", func() {
				Expect(testSink.Errors).To(Equal([]error{err}))
			})
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				defer func() {
					fatalErr = recover()
				}()

				logger.Fatal(action, err)
			})

			TestCommonLogFeatures(lager.FATAL)

			It("data contains error message", func() {
				Expect(testSink.Logs()[0].Data["error"]).To(Equal(err.Error()))
			})

			It("data contains stack trace", func() {
				Expect(testSink.Logs()[0].Data["trace"]).NotTo(BeEmpty())
			})

			It("panics with the provided error", func() {
				Expect(fatalErr).To(Equal(err))
			})

			It("retains the original error values", func() {
				Expect(testSink.Errors).To(Equal([]error{err}))
			})
		})

		Context("with no error", func() {
			BeforeEach(func() {
				defer func() {
					fatalErr = recover()
				}()

				logger.Fatal(action, nil)
			})

			TestCommonLogFeatures(lager.FATAL)

			It("does not contain the error message", func() {
				Expect(testSink.Logs()[0].Data).NotTo(HaveKey("error"))
			})

			It("data contains stack trace", func() {
				Expect(testSink.Logs()[0].Data["trace"]).NotTo(BeEmpty())
			})

			It("panics with the provided error (i.e. nil)", func() {
				Expect(fatalErr).To(BeNil())
			})
		})

	})

	Describe("WithTraceInfo", func() {
		var req *http.Request

		BeforeEach(func() {
			var err error
			req, err = http.NewRequest("GET", "/foo", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when request does not contain trace id", func() {
			It("does not set trace and span id", func() {
				logger = logger.WithTraceInfo(req)
				logger.Info("test-log")

				log := testSink.Logs()[0]

				Expect(log.Data).To(BeEmpty())
				Expect(log.Data).To(BeEmpty())
			})
		})

		Context("when request contains trace id", func() {
			It("sets trace and span id", func() {
				req.Header.Set("X-Vcap-Request-Id", "7f461654-74d1-1ee5-8367-77d85df2cdab")

				logger = logger.WithTraceInfo(req)
				logger.Info("test-log")

				log := testSink.Logs()[0]

				Expect(log.Data["trace-id"]).To(Equal("7f46165474d11ee5836777d85df2cdab"))
				Expect(log.Data["span-id"]).NotTo(BeEmpty())
			})

			It("generates new span id", func() {
				req.Header.Set("X-Vcap-Request-Id", "7f461654-74d1-1ee5-8367-77d85df2cdab")

				logger = logger.WithTraceInfo(req)
				logger.Info("test-log")

				log1 := testSink.Logs()[0]

				Expect(log1.Data["trace-id"]).To(Equal("7f46165474d11ee5836777d85df2cdab"))
				Expect(log1.Data["span-id"]).NotTo(BeEmpty())

				logger = logger.WithTraceInfo(req)
				logger.Info("test-log")

				log2 := testSink.Logs()[1]

				Expect(log2.Data["trace-id"]).To(Equal("7f46165474d11ee5836777d85df2cdab"))
				Expect(log2.Data["span-id"]).NotTo(BeEmpty())
				Expect(log2.Data["span-id"]).NotTo(Equal(log1.Data["span-id"]))
			})
		})

		Context("when request contains invalid trace id", func() {
			It("does not set trace and span id", func() {
				req.Header.Set("X-Vcap-Request-Id", "invalid-request-id")

				logger = logger.WithTraceInfo(req)
				logger.Info("test-log")

				log := testSink.Logs()[0]

				Expect(log.Data).To(BeEmpty())
				Expect(log.Data).To(BeEmpty())
			})
		})
	})
})
