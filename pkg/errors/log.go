package errors

import "github.com/fhofherr/golf/log"

// Log logs the passed error to logger. Does nothing if logger or error is
// nil.
func Log(logger log.Logger, err error) {
	if logger == nil || err == nil {
		return
	}
	// TODO (fhofherr): check if err is of type Error and construct a better log entry
	log.Log(logger, "level", "error", "error", err)
}

// LogFunc calles the passed function f. Any error returned by fis logged using
// Log.
func LogFunc(logger log.Logger, f func() error) {
	if f == nil {
		return
	}
	Log(logger, f())
}
