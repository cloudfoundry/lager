package lager_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"runtime"

	"testing"
)

const MaxThreads = 100
var _ = BeforeSuite(func() {
	runtime.GOMAXPROCS(MaxThreads)
})

func TestLager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lager Suite")
}
