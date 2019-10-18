package errors

import (
	"errors"

	"github.com/fhofherr/golf/log"
)

// Log logs the passed error to logger. Does nothing if logger or error is
// nil.
func Log(logger log.Logger, err error) {
	var acpErr *Error

	if logger == nil || err == nil {
		return
	}

	if !errors.As(err, &acpErr) {
		log.Log(logger, "level", "error", "message", err.Error(), "error", err)
		return
	}
	log.Log(logger,
		"level", "error",
		"message", acpErr.Msg,
		"trace", acpErr.Trace(),
		"error", err)
}

// LogFunc calles the passed function f. Any error returned by fis logged using
// Log.
func LogFunc(logger log.Logger, f func() error) {
	if f == nil {
		return
	}
	Log(logger, f())
}
