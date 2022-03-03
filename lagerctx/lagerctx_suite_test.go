package lagerctx_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLagerctx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lagerctx Suite")
}
