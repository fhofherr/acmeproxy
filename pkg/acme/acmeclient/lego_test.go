package acmeclient_test

import (
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/golf/log"
	legolog "github.com/go-acme/lego/log"
	"github.com/stretchr/testify/assert"
)

func TestInitializeLego(t *testing.T) {
	original := legolog.Logger
	defer func() {
		legolog.Logger = original
	}()

	logger := &log.TestLogger{}
	acmeclient.InitializeLego(logger)
	assert.NotEqual(t, original, legolog.Logger)

	legolog.Logger.Print("some message")
	logger.AssertHasMatchingLogEntries(t, 1, func(e log.TestLogEntry) bool {
		return e["message"] == "some message"
	})

	anotherLogger := &log.TestLogger{}
	acmeclient.InitializeLego(anotherLogger)
	assert.NotEqual(t, anotherLogger, legolog.Logger)
}
