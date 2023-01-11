package lagerctx_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagerctx"
	"code.cloudfoundry.org/lager/v3/lagertest"
)

var _ = Describe("Lager Context", func() {
	It("can store loggers inside contexts", func() {
		l := lagertest.NewTestLogger("lagerctx")
		ctx := lagerctx.NewContext(context.Background(), l)

		logger := lagerctx.FromContext(ctx)
		logger.Info("from-a-context")

		Expect(l.LogMessages()).To(HaveLen(1))
	})

	It("can add a session to the logger in the context", func() {
		l := lagertest.NewTestLogger("lagerctx")
		ctx := lagerctx.NewContext(context.Background(), l)

		logger := lagerctx.WithSession(ctx, "new-session", lager.Data{
			"bespoke-data": "",
		})
		logger.Info("from-a-context")

		Expect(l).To(gbytes.Say("new-session"))
		Expect(l).To(gbytes.Say("bespoke-data"))
	})

	It("can add data to the logger in the context", func() {
		l := lagertest.NewTestLogger("lagerctx")
		ctx := lagerctx.NewContext(context.Background(), l)

		logger := lagerctx.WithData(ctx, lager.Data{
			"bespoke-data": "",
		})
		logger.Info("from-a-context")

		Expect(l).To(gbytes.Say("bespoke-data"))
	})

	It("will be fine if there is no logger in the context", func() {
		logger := lagerctx.FromContext(context.Background())
		logger.Info("from-a-context")
	})
})
