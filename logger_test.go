package lager_test

import (
	"fmt"

	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var logger lager.Logger
	var testSink *lager.TestLogger

	var component = "my-component"
	var task = "my-task"
	var action = "my-action"
	var description = "my-description"

	BeforeEach(func() {
		logger = lager.NewLogger(component)
		testSink = lager.NewTestLogger()
		logger.RegisterSink(testSink)
	})

	Describe("Debug", func() {
		BeforeEach(func() {
			logger.Debug(task, action, description, lager.Data{"foo": "bar"})
		})
		It("writes the proper log format", func() {
			logs := testSink.Logs()

			Ω(logs).Should(HaveLen(1))

			Ω(logs[0].LogLevel).Should(Equal(lager.DEBUG))
			Ω(logs[0].Source).Should(Equal(component))
			Ω(logs[0].Message).Should(Equal(fmt.Sprintf("%s.%s.%s", component, task, action)))
			Ω(logs[0].Data).Should(Equal(lager.Data{"description": description, "foo": "bar"}))
		})
	})
})
