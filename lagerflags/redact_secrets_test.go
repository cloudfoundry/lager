package lagerflags_test

import (
	"flag"

	"code.cloudfoundry.org/lager/lagerflags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RedactSecrets", func() {

	Context("RedactPatterns FlagSet", func() {
		var flagSet *flag.FlagSet
		var pattern lagerflags.RedactPatterns

		BeforeEach(func() {
			pattern = nil
			flagSet = flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.Usage = func() {}
			flagSet.SetOutput(nopWriter{})
			flagSet.Var(
				&pattern,
				"pattern",
				`redaction pattern`,
			)
		})

		It("parses successfully when no flag is supplied", func() {
			Expect(flagSet.Parse([]string{})).To(Succeed())
			Expect(pattern).To(BeNil())
		})

		It("parses successfully when one flag is supplied", func() {
			Expect(flagSet.Parse([]string{"-pattern", "foo"})).To(Succeed())
			Expect(pattern).To(Equal(lagerflags.RedactPatterns{"foo"}))
		})

		It("parses successfully when flags are supplied", func() {
			Expect(flagSet.Parse([]string{
				"-pattern",
				"one",
				"-pattern",
				"two",
				"-pattern",
				"three",
			})).To(Succeed())
			Expect(pattern).To(ContainElement("one"))
			Expect(pattern).To(ContainElement("two"))
			Expect(pattern).To(ContainElement("three"))
		})
	})
})
