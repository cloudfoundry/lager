//go:build go1.21

package lager_test

import (
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
        "fmt"
	"log/slog"
	"strconv"
	"strings"
	"testing/slogtest"
	"time"
)

var _ = Describe("NewHandler", func() {
	var (
		s *lagertest.TestSink
		l lager.Logger
		h slog.Handler
	)

	BeforeEach(func() {
		s = lagertest.NewTestSink()
		l = lager.NewLogger("test")
		l.RegisterSink(s)

		h = lager.NewHandler(l)
	})

	It("logs a message", func() {
		slog.New(h).Info("foo", "bar", "baz")
		logs := s.Logs()
		Expect(logs).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Source":  Equal("test"),
			"Message": Equal("test.foo"),
			"Data": SatisfyAll(
				HaveLen(1),
				HaveKeyWithValue("bar", "baz"),
			),
			"LogLevel": Equal(lager.INFO),
		})))
	})

	It("logs a debug message", func() {
		slog.New(h).Debug("foo", "bar", 3, slog.Int("baz", 42))
		logs := s.Logs()
		Expect(logs).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Source":  Equal("test"),
			"Message": Equal("test.foo"),
			"Data": SatisfyAll(
				HaveLen(2),
				HaveKeyWithValue("bar", float64(3)),
				HaveKeyWithValue("baz", float64(42)),
			),
			"LogLevel": Equal(lager.DEBUG),
		})))
	})

	It("logs an error message", func() {
		slog.New(h).Error("foo", "error", fmt.Errorf("boom"))
		logs := s.Logs()
		Expect(logs).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Source":  Equal("test"),
			"Message": Equal("test.foo"),
			"Data": SatisfyAll(
				HaveLen(1),
				HaveKeyWithValue("error", "boom"),
			),
			"LogLevel": Equal(lager.ERROR),
		})))
	})

	It("behaves like a slog.NewHandler", func() {
		results := func() (result []map[string]any) {
			for _, l := range s.Logs() {
				d := l.Data

				t := parseTimestamp(l.Timestamp)
				if !t.IsZero() {
					d["time"] = t
				}

				d["level"] = l.LogLevel
				d["msg"] = strings.TrimPrefix(l.Message, "test.")
				result = append(result, d)
			}
			return result
		}

		Expect(slogtest.TestHandler(h, results)).To(Succeed())
	})
})

// parseTimestamp turns a lager timestamp back into a time.Time
// with a special case for the formatting of time.Time{} because
// there is a test that check time.Time{} is ignored as a time value
func parseTimestamp(input string) time.Time {
	GinkgoHelper()

	// This is what time.Time{} gets formatted as
	if input == "-6795364578.871345520" {
		return time.Time{}
	}

	f64, err := strconv.ParseFloat(input, 64)
	Expect(err).NotTo(HaveOccurred())

	secs := int64(f64)
	nanos := int64((f64 - float64(secs)) * 1e9)
	return time.Unix(secs, nanos)
}
