package errors_test

import (
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/golf/log"
)

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
				return e["level"] == "error" && errors.Match(errors.New(errors.Op("test-op"), "something failed"), err)
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
