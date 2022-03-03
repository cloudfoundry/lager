package chug_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestChug(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chug Suite")
}
