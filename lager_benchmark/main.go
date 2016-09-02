package main

import (
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/lager"
)

type SimpleStruct struct {
	AnInt         int
	AString       string
	AnotherString string
	AnotherInt    uint64
}

type ComplexStruct struct {
	AnIntArray          []int
	AString             string
	ASimpleStruct       SimpleStruct
	AStringArray        []string
	AnotherSimpleStruct SimpleStruct
}

func main() {
	logFile, err := os.Create("log_file.log")
	if err != nil {
		fmt.Printf("unable to open file: %#v\n", err)
		os.Exit(1)
	}

	defer logFile.Close()

	logger := lager.NewLogger("lager_benchmark")
	logger.RegisterSink(lager.NewWriterSink(logFile, lager.INFO))
	logger.Info("started")
	defer logger.Info("exited")

	simple1 := SimpleStruct{-4, "quack", "moo", 23}
	simple2 := SimpleStruct{9, "jake", "finn", 148}

	complex := ComplexStruct{
		[]int{913, 4715, -2990, 8},
		"prosciutto",
		simple1,
		[]string{"people", "are strange", "when you're a stranger"},
		simple2,
	}

	var accumulator int64
	nIterations := 1000000

	for i := 0; i < nIterations; i++ {
		start := time.Now()

		// logger.Info("simple-struct", lager.Data{"a-key": "sandwich", "payload": simple1})
		// logger.Info("another-simple-struct", lager.Data{"a-key": "pastrami", "payload": simple2})
		logger.Info("a-complex-struct", lager.Data{"a-key": "gruyere", "payload": complex})

		finished := time.Since(start)
		accumulator += finished.Nanoseconds()
	}

	fmt.Printf("%f\n", float64(accumulator)/float64(nIterations))
}
