package main

import (
	"errors"
	"flag"

	"code.cloudfoundry.org/lager/v2"
	"code.cloudfoundry.org/lager/v2/lagerflags"
)

func main() {
	lagerflags.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := lagerflags.New("cf-lager-integration")

	logger.Debug("component-does-action", lager.Data{"debug-detail": "foo"})
	logger.Info("another-component-action", lager.Data{"info-detail": "bar"})
	logger.Error("component-failed-something", errors.New("error"), lager.Data{"error-detail": "baz"})
	logger.Fatal("component-failed-badly", errors.New("fatal"), lager.Data{"fatal-detail": "quux"})
}
