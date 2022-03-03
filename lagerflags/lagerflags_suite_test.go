package lagerflags_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lager Flags Suite")
}
