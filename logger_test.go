package lager_test

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var logger lager.Logger
	var testSink *lager.TestSink

	var component = "my-component"
	var task = "my-task"
	var action = "my-action"
	var description = "my-description"
	var logData = lager.Data{"foo": "bar"}
	var logs []lager.LogFormat

	BeforeEach(func() {
		logger = lager.NewLogger(component)
		testSink = lager.NewTestSink()
		logger.RegisterSink(testSink)
		logs = nil
	})

	var TestForCommonLogFeatures = func() {
		var log lager.LogFormat

		BeforeEach(func() {
			log = logs[0]
		})

		It("writes a log to the sink", func() {
			Ω(logs).Should(HaveLen(1))
		})

		It("records the source component", func() {
			Ω(log.Source).Should(Equal(component))
		})

		It("outputs a properly-formatted message", func() {
			Ω(log.Message).Should(Equal(fmt.Sprintf("%s.%s.%s", component, task, action)))
		})

		It("has a timestamp", func() {
			expectedTime := float64(time.Now().UnixNano()) / 1e9
			parsedTimestamp, err := strconv.ParseFloat(log.Timestamp, 64)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(parsedTimestamp).Should(BeNumerically("~", expectedTime, 1.0))
		})

		It("data contains the description", func() {
			Ω(log.Data["description"]).Should(Equal(description))
		})

		It("data contains custom user data", func() {
			Ω(log.Data["foo"]).Should(Equal("bar"))
		})
	}

	Describe("Debug", func() {
		BeforeEach(func() {
			logger.Debug(task, action, description, logData)
			logs = testSink.Logs()
		})

		TestForCommonLogFeatures()

		It("sets the proper output level", func() {
			Ω(logs[0].LogLevel).Should(Equal(lager.DEBUG))
		})
	})

	Describe("Info", func() {
		BeforeEach(func() {
			logger.Info(task, action, description, logData)
			logs = testSink.Logs()
		})

		TestForCommonLogFeatures()

		It("sets the proper output level", func() {
			Ω(logs[0].LogLevel).Should(Equal(lager.INFO))
		})

	})

	Describe("Error", func() {
		var err = errors.New("oh noes!")

		BeforeEach(func() {
			logger.Error(task, action, description, err, logData)
			logs = testSink.Logs()
		})

		TestForCommonLogFeatures()

		It("sets the proper output level", func() {
			Ω(logs[0].LogLevel).Should(Equal(lager.ERROR))
		})

		It("data contains error message", func() {
			Ω(logs[0].Data["error"]).Should(Equal(err.Error()))
		})
	})

	Describe("Fatal", func() {
		var err = errors.New("oh noes!")
		var fatalErr interface{}
		BeforeEach(func() {
			defer func() {
				fatalErr = recover()
				logs = testSink.Logs()
			}()

			logger.Fatal(task, action, description, err, logData)
		})

		TestForCommonLogFeatures()

		It("sets the proper output level", func() {
			Ω(logs[0].LogLevel).Should(Equal(lager.FATAL))
		})

		It("data contains error message", func() {
			Ω(logs[0].Data["error"]).Should(Equal(err.Error()))
		})

		It("data contains stack trace", func() {
			Ω(logs[0].Data["trace"]).ShouldNot(BeEmpty())
		})

		It("panics with the provided error", func() {
			Ω(fatalErr).Should(Equal(err))
		})
	})
})
