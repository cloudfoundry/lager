package lagerflags_test

import (
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
)

// TODO: Allow sink output to be redirected to a dependency injected
// io.Writer
func replaceStdoutWithBuf() (*gbytes.Buffer, *os.File) {
	buf := gbytes.NewBuffer()
	readPipe, writePipe, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = writePipe
	go func() {
		p := make([]byte, 20)
		for {
			n, err := readPipe.Read(p)
			if err == io.EOF {
				break
			}

			_, err = buf.Write(p[:n])
			if err != nil {
				return
			}
		}
	}()
	return buf, origStdout
}

var _ = Describe("Lagerflags", func() {
	Context("when parsing flags", func() {
		var flagSet *flag.FlagSet

		BeforeEach(func() {
			flagSet = flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.SetOutput(ioutil.Discard)
			lagerflags.AddFlags(flagSet)
		})

		Describe("ConfigFromFlags", func() {
			It("creates the correct Lager config from parsed flags", func() {
				err := flagSet.Parse([]string{"-logLevel", "debug", "-redactSecrets", "-timeFormat", "unix-epoch"})
				Expect(err).NotTo(HaveOccurred())

				c := lagerflags.ConfigFromFlags()
				Expect(c).To(Equal(lagerflags.LagerConfig{
					LogLevel:            string(lagerflags.DEBUG),
					RedactSecrets:       true,
					TimeFormat:          lagerflags.FormatUnixEpoch,
					MaxDataStringLength: 0,
				}))
			})
		})

		Describe("New", func() {
			It("creates a logger that respects the log level from parsed flags", func() {
				err := flagSet.Parse([]string{"-logLevel", "error"})
				Expect(err).NotTo(HaveOccurred())

				buf, origStdout := replaceStdoutWithBuf()
				defer func() {
					os.Stdout = origStdout
				}()

				logger, _ := lagerflags.New("test")
				logger.Info("hello")
				Consistently(buf).ShouldNot(gbytes.Say("hello"))
				logger.Error("foo", errors.New("kaboom"))
				Eventually(buf).Should(gbytes.Say("kaboom"))
			})

			It("creates a logger that respects the time format settings from parsed flags", func() {
				err := flagSet.Parse([]string{"-timeFormat", "rfc3339"})
				Expect(err).NotTo(HaveOccurred())

				buf, origStdout := replaceStdoutWithBuf()
				defer func() {
					os.Stdout = origStdout
				}()

				logger, _ := lagerflags.New("test")
				logger.Info("hello")
				Eventually(buf).Should(gbytes.Say(`"timestamp":"(\d+)-(\d+)-(\d+)[Tt](\d+):(\d+):(\d+).(\d+)Z`))
			})
		})
	})

	Describe("NewFromConfig", func() {
		It("creates a logger that respects the log level", func() {
			buf, origStdout := replaceStdoutWithBuf()
			defer func() {
				os.Stdout = origStdout
			}()

			logger, _ := lagerflags.NewFromConfig("test", lagerflags.LagerConfig{
				LogLevel: lagerflags.ERROR,
			})

			logger.Info("hello")
			Consistently(buf).ShouldNot(gbytes.Say("hello"))
			logger.Error("foo", errors.New("kaboom"))
			Eventually(buf).Should(gbytes.Say("kaboom"))
		})

		It("creates a logger that respects the time format settings", func() {
			buf, origStdout := replaceStdoutWithBuf()
			defer func() {
				os.Stdout = origStdout
			}()

			logger, _ := lagerflags.NewFromConfig("test", lagerflags.LagerConfig{
				LogLevel:   lagerflags.INFO,
				TimeFormat: lagerflags.FormatRFC3339,
			})

			logger.Info("hello")
			Eventually(buf).Should(gbytes.Say(`"timestamp":"(\d+)-(\d+)-(\d+)[Tt](\d+):(\d+):(\d+).(\d+)Z`))
		})

		It("creates a logger that redacts secrets", func() {
			buf, origStdout := replaceStdoutWithBuf()
			defer func() {
				os.Stdout = origStdout
			}()

			logger, _ := lagerflags.NewFromConfig("test", lagerflags.LagerConfig{
				LogLevel:      lagerflags.INFO,
				RedactSecrets: true,
			})

			logger.Info("hello", lager.Data{"password": "data"})
			Eventually(buf).Should(gbytes.Say(`"password":"\*REDACTED\*"`))
		})

		It("creates a logger that redacts secrets using the supplied redaction regex", func() {
			buf, origStdout := replaceStdoutWithBuf()
			defer func() {
				os.Stdout = origStdout
			}()

			logger, _ := lagerflags.NewFromConfig("test", lagerflags.LagerConfig{
				LogLevel:       lagerflags.INFO,
				RedactSecrets:  true,
				RedactPatterns: []string{"bar"},
			})

			logger.Info("hello", lager.Data{"foo": "bar"})
			Eventually(buf).Should(gbytes.Say(`"foo":"\*REDACTED\*"`))
		})

		It("creates a logger that truncates long strings", func() {
			buf, origStdout := replaceStdoutWithBuf()
			defer func() {
				os.Stdout = origStdout
			}()

			logger, _ := lagerflags.NewFromConfig("test", lagerflags.LagerConfig{
				LogLevel:            lagerflags.INFO,
				MaxDataStringLength: 20,
			})

			logger.Info("hello", lager.Data{"password": strings.Repeat("a", 25)})
			Eventually(buf).Should(gbytes.Say(`"password":"aaaaaaaa-\(truncated\)"`))
		})

		It("panics if the log level is unknown", func() {
			Expect(func() {
				_, _ = lagerflags.NewFromConfig("test", lagerflags.LagerConfig{
					LogLevel: "foo",
				})
			}).To(Panic())
		})
	})
})
