//go:build go1.21

package lager_test

import (
	"bytes"
	"code.cloudfoundry.org/lager/v3"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"log/slog"
)

var _ = Describe("NewSlogSink", func() {
	var (
		buf    bytes.Buffer
		logger lager.Logger
	)

	matchTimestamp := MatchRegexp(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{5,9}Z$`)

	parsedLogMessage := func() (receiver map[string]any) {
		Expect(json.Unmarshal(buf.Bytes(), &receiver)).To(Succeed())
		return
	}

	BeforeEach(func() {
		buf = bytes.Buffer{}
		logger = lager.NewLogger("fake-component")
		logger.RegisterSink(lager.NewSlogSink(slog.New(slog.NewJSONHandler(&buf, nil))))
	})

	It("logs Info()", func() {
		logger.Info("fake-info", lager.Data{"foo": "bar"})

		Expect(parsedLogMessage()).To(MatchAllKeys(Keys{
			"time":  matchTimestamp,
			"level": Equal("INFO"),
			"msg":   Equal("fake-component.fake-info"),
			"foo":   Equal("bar"),
		}))
	})

	It("logs Debug()", func() {
		logger.Debug("fake-debug", lager.Data{"foo": "bar"})

		Expect(parsedLogMessage()).To(MatchAllKeys(Keys{
			"time":  matchTimestamp,
			"level": Equal("DEBUG"),
			"msg":   Equal("fake-component.fake-debug"),
			"foo":   Equal("bar"),
		}))
	})

	It("logs Error()", func() {
		logger.Error("fake-error", fmt.Errorf("boom"), lager.Data{"foo": "bar"})

		Expect(parsedLogMessage()).To(MatchAllKeys(Keys{
			"time":  matchTimestamp,
			"error": Equal("boom"),
			"level": Equal("ERROR"),
			"msg":   Equal("fake-component.fake-error"),
			"foo":   Equal("bar"),
		}))
	})

	It("logs Fatal()", func() {
		Expect(func() {
			logger.Fatal("fake-fatal", fmt.Errorf("boom"), lager.Data{"foo": "bar"})
		}).To(Panic())

		Expect(parsedLogMessage()).To(MatchAllKeys(Keys{
			"time":  matchTimestamp,
			"error": Equal("boom"),
			"level": Equal("ERROR"),
			"msg":   Equal("fake-component.fake-fatal"),
			"foo":   Equal("bar"),
			"trace": ContainSubstring(`code.cloudfoundry.org/lager/v3.(*logger).Fatal`),
		}))
	})
})
