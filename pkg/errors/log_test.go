package errors_test

import (
	"fmt"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/golf/log"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	tests := []struct {
		name     string
		logger   *log.TestLogger
		err      error
		nEntries int
		level    string
		message  string
		trace    []errors.Op
	}{
		{
			name:   "nil error",
			logger: &log.TestLogger{},
		},
		{
			name:     "plain error",
			logger:   &log.TestLogger{},
			err:      fmt.Errorf("some error"),
			nEntries: 1,
			level:    "error",
			message:  fmt.Errorf("some error").Error(),
		},
		{
			name:     "custom error",
			logger:   &log.TestLogger{},
			err:      errors.New(errors.Op("some op"), "some error"),
			nEntries: 1,
			level:    "error",
			message:  "some error",
			trace: []errors.Op{
				errors.Op("some op"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errors.Log(tt.logger, tt.err)
			tt.logger.AssertHasMatchingLogEntries(t, tt.nEntries, func(e log.TestLogEntry) bool {
				traceMatches := true
				if tt.trace != nil {
					traceMatches = assert.ObjectsAreEqual(tt.trace, e["trace"])
				}
				return e["level"] == tt.level && e["message"] == tt.message && traceMatches
			})
		})
	}
}

func TestLogFunc(t *testing.T) {
	tests := []struct {
		name     string
		f        func() error
		logger   *log.TestLogger
		nEntries int
		pred     func(log.TestLogEntry) bool
	}{
		{
			name:   "nil closer",
			logger: &log.TestLogger{},
		},
		{
			name: "nil logger",
			f: func() error {
				return nil
			},
		},
		{
			name: "function is successful",
			f: func() error {
				return nil
			},
			logger: &log.TestLogger{},
		},
		{
			name: "function fails with custom error",
			f: func() error {
				return errors.New(errors.Op("test-op"), "something failed")
			},
			logger:   &log.TestLogger{},
			nEntries: 1,
			pred: func(e log.TestLogEntry) bool {
				maybeErr, found := e["error"]
				if !found {
					return false
				}
				err, ok := maybeErr.(*errors.Error)
				if !ok {
					return false
				}
				return e["level"] == "error" && errors.Is(err, errors.New(errors.Op("test-op"), "something failed"))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.pred == nil {
				tt.pred = func(log.TestLogEntry) bool {
					return true
				}
			}
			errors.LogFunc(tt.logger, tt.f)
			if tt.logger == nil {
				// We are happy if nothing happened
				return
			}
			tt.logger.AssertHasMatchingLogEntries(t, tt.nEntries, tt.pred)
		})
	}
}
