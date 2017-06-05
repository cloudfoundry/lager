package lager_test

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {

	Context("when used as a Logger", func() {
		var (
			logger   lager.Context
			testSink *lagertest.TestSink
		)

		BeforeEach(func() {
			logger = lager.NewContext(context.TODO(), lager.NewLogger("logger-context-test"))
			testSink = lagertest.NewTestSink()
			logger.RegisterSink(testSink)
		})

		It("should be castable to lager.Logger", func() {
			c, ok := logger.(context.Context)
			Expect(ok).To(Equal(true))
			Expect(c).NotTo(BeNil())
		})

		Context("when written to", func() {
			BeforeEach(func() {
				logger.Info("some info")
			})

			It("should log", func() {
				log := testSink.Logs()[0]
				Expect(log.Message).To(Equal("logger-context-test.some info"))
			})
		})
	})

	Context("when used as a Context", func() {
		var (
			ctx lager.Context
		)

		BeforeEach(func() {
			ctx = lager.NewContext(context.TODO(), lager.NewLogger("context-context-test"))
		})

		It("should be castable to context.Context", func() {
			c, ok := ctx.(context.Context)
			Expect(ok).To(Equal(true))
			Expect(c).NotTo(BeNil())
		})

		Context("for cancellation", func() {
			var (
				cancel context.CancelFunc
			)

			BeforeEach(func() {
				ctx, cancel = lager.WithCancel(ctx)
			})

			It("should cancel as expected", func() {
				Expect(ctx.Done()).NotTo(BeNil())

				select {
				case x := <-ctx.Done():
					Fail(fmt.Sprintf("Unexpected signal %#v", x))
				default:
				}

				cancel()
				time.Sleep(100 * time.Millisecond) // let cancellation propagate

				select {
				case <-ctx.Done():
				default:
					Fail(fmt.Sprintf("Signal expected"))
				}

				Expect(ctx.Err()).To(Equal(context.Canceled))
			})
		})

		Context("for timeout", func() {
			BeforeEach(func() {
				ctx, _ = lager.WithTimeout(ctx, 50*time.Millisecond)
			})

			It("should timeout as expected", func() {
				select {
				case <-time.After(time.Second):
					Fail("Context should have timed out")
				case <-ctx.Done():
				}
				Expect(ctx.Err()).To(Equal(context.DeadlineExceeded))
			})
		})
	})
})
