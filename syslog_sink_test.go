package lager_test

import (
	"bufio"
	"io"
	"net"
	"time"
	. "github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//taken from http://golang.org/src/pkg/log/syslog/syslog_test.go
func runStreamSyslog(slow bool, listener net.Listener, results chan<- string) {
	for {
		var connection net.Conn
		var err error
		if connection, err = listener.Accept(); err != nil {
			return
		}
		if slow {
			time.Sleep(100 * time.Minute)
		}
		go func(connection net.Conn) {
			connection.SetReadDeadline(time.Now().Add(5 * time.Second))
			buffer := bufio.NewReader(connection)
			for ct := 1; ct&7 != 0; ct++ {
				s, err := buffer.ReadString('\n')
				if err != nil {
					break
				}
				results <- s
			}
			connection.Close()
		}(connection)
	}
}

func startServer(slow bool, results chan<- string) (addr string, sock io.Closer) {
	listenerAddress := "127.0.0.1:0"
	transport := "tcp"

	listener, err := net.Listen(transport, listenerAddress)
	Ω(err).ShouldNot(HaveOccurred())

	addr = listener.Addr().String()
	sock = listener
	go func() {
		runStreamSyslog(slow, listener, results)
	}()

	return addr, listener
}

var _ = Describe("SyslogSink", func() {
	var sink Sink
	var results chan string
	var serverAddress string
	var listener io.Closer

	Context("with a working syslog server", func() {
		BeforeEach(func() {
			var err error

			results = make(chan string)
			serverAddress, listener = startServer(false, results)
			sink, err = NewSyslogSink("tcp", serverAddress, "my-tag", INFO)
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			listener.Close()
		})

		Context("when the logging level is above the minimum level", func() {
			It("should log to syslog", func() {
				sink.Log(INFO, []byte("hello"))

				Eventually(results).Should(Receive(MatchRegexp(`my-tag\[\d+\]: hello`)))
			})
		})

		Context("when the logging level is below the minimum level", func() {
			It("should not log to syslog", func() {
				sink.Log(DEBUG, []byte("hello"))

				Consistently(results).ShouldNot(Receive())
			})
		})

		Describe("supporting different log levels", func() {
			BeforeEach(func() {
				var err error
				sink, err = NewSyslogSink("tcp", serverAddress, "my-tag", DEBUG)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("should emit an Debug message when given a DEBUG log", func() {
				sink.Log(DEBUG, []byte("hello"))

				Eventually(results).Should(Receive(ContainSubstring("<31>")))
			})

			It("should emit an Info message when given an INFO log", func() {
				sink.Log(INFO, []byte("hello"))

				Eventually(results).Should(Receive(ContainSubstring("<30>")))
			})

			It("should emit an Err message when given an ERROR log", func() {
				sink.Log(ERROR, []byte("hello"))

				Eventually(results).Should(Receive(ContainSubstring("<27>")))
			})

			It("should emit an Emerg message when given a FATAL log", func() {
				sink.Log(FATAL, []byte("hello"))

				Eventually(results).Should(Receive(ContainSubstring("<24>")))
			})
		})
	})

	Context("with a slow syslog server", func() {
		BeforeEach(func() {
			var err error

			results = make(chan string)
			serverAddress, listener = startServer(true, results)
			sink, err = NewSyslogSink("tcp", serverAddress, "my-tag", INFO)
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			listener.Close()
		})

		It("should never block", func(done Done) {
			for i := 0; i < 1000000; i++ { //do this alot to fill the buffer deep in the bowels of syslog
				sink.Log(INFO, []byte("hello"))
			}
			close(done)
		})
	})

	Context("with no server", func() {
		It("should return an error, but not get stuck", func() {
			var err error
			sink, err = NewSyslogSink("tcp", "127.0.0.1:12382", "my-tag", INFO)
			Ω(err).Should(HaveOccurred())
		})
	})
})
