package errors

import "strings"

// Collection contains multiple errors that occurred during an operation.
type Collection []error

// Append add err to the collection c, if err is not nil.
//
// It does so by calling Wrap with err and args. If the return value of Wrap is
// not nil it gets appended to c.
func Append(c Collection, err error, args ...interface{}) Collection {
	if err == nil {
		return c
	}
	return append(c, Wrap(err, args...))
}

// Error formats the collection c as a string.
//
// If c is empty the empty string is returned. If c has size one the return
// value of the Error method of the only entry is returned. Otherwise the return
// value of Error of each entry in c is prefixed with '* ' and added on a
// separate line to the returned string.
func (c Collection) Error() string {
	if c == nil || len(c) == 0 {
		return ""
	}
	if len(c) == 1 {
		return c[0].Error()
	}

	var buf strings.Builder
	for _, err := range c {
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("* ")
		buf.WriteString(err.Error())
	}
	return buf.String()
}

// ErrorOrNil returns c if c is not nil and has a size greater 0. Otherwise
// it returns nil.
func (c Collection) ErrorOrNil() error {
	if c == nil || len(c) == 0 {
		return nil
	}
	return c
}
