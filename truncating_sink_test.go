package lager_test

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const longString = "aaaaaaaaaaaaaaaaaaaaaaaaa"

var _ = Describe("TruncatingSink", func() {
	var (
		sink     lager.Sink
		testSink *lagertest.TestSink
	)

	type dummyStruct struct {
		A string
	}

	BeforeEach(func() {
		testSink = lagertest.NewTestSink()

		sink = lager.NewTruncatingSink(testSink, 20)
	})
	Context("when given data that has only short strings", func() {
		BeforeEach(func() {
			sink.Log(lager.LogFormat{
				LogLevel: lager.INFO,
				Message:  "hello world",
				Data:     lager.Data{"foo": "bar", "dummy": dummyStruct{A: "abcd"}},
			})
		})
		It("writes the data to the given sink without truncating any strings", func() {
			Expect(testSink.Buffer().Contents()).To(
				MatchJSON(`{"timestamp":"","log_level":1,"source":"","message":"hello world","data":{"foo":"bar","dummy":{"A":"abcd"}}}`),
			)
		})
	})
	Context("when given data that includes strings that exceed the configured max length", func() {
		BeforeEach(func() {
			sink.Log(lager.LogFormat{
				LogLevel: lager.INFO,
				Message:  "hello world",
				Data:     lager.Data{"foo": longString, "dummy": dummyStruct{A: longString}},
			})
		})
		It("truncates the data and writes to the given sink", func() {
			Expect(testSink.Buffer().Contents()).To(
				MatchJSON(`{"timestamp":"","log_level":1,"source":"","message":"hello world","data":{"foo":"aaaaaaaa-(truncated)","dummy":{"A":"aaaaaaaa-(truncated)"}}}`),
			)
		})
	})
})
