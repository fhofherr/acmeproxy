package errors

import (
	"errors"
	"reflect"
	"strings"
)

// Is is a convenience wrapper for the Go standard library errors/Is function.
// It allows all of acmeproxy to use this errors package instead of importing
// one of the two packages under an alias.
//
// See https://godoc.org/errors#Is for documentation.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As is a convenience wrapper for the Go standard library errors/As function.
// It allows all of acmeproxy to use this errors package instead of importing
// one of the two packages under an alias.
//
// See https://godoc.org/errors#As for documentation.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Op encapsulates the name of an operation.
//
// It should contain the package and function name.
type Op string

// Kind categorizes the nature of an error.
//
// Errors of different kind may be treated with different severities. Code
// may decide to use the Kind of an error to translate it into a response for
// a client.
type Kind int

const (
	// Unspecified is the default value for the an Error's Kind if nothing else
	// was specified. Code should treat errors of Kind Unspecified as fatal.
	Unspecified Kind = iota

	// NotFound shows that an error was returned because something could not
	// be found. Callers may choose to continue with a fallback for the thing
	// that could not be found, or abort the operation.
	NotFound

	// InvalidArgument shows that an opperation was called with one or more
	// invalid arguments. This usually indicates a client error. An HTTP API
	// might return a status 400 Bad Request response.
	InvalidArgument
)

func (k Kind) String() string {
	switch k {
	case Unspecified:
		return ""
	case NotFound:
		return "not found"
	case InvalidArgument:
		return "invalid argument"
	default:
		return "unknown kind"
	}
}

// Error represents an error within acmeproxy.
//
// The fields of Error provide additional detail about the error. Any of Error's
// fields may be left unset if it is not applicable for the error.
//
// The field Err allows the error to reference another error. This allows to
// build a chain of errors. If the referenced error is an instance of Error
// itself the Op fields of the errors in the chain can be used to build a trace
// which should lead to the root cause of the error.
type Error struct {
	Op   Op
	Kind Kind

	// Msg is a message detailing the error further. It should follow the
	// usual Go conventions for messages in errors, i.e. start with a lower-case
	// letter and be relatively short but meaningful,
	Msg string

	// Err references an error which led to this error being returned. Err
	// is used to build a trace of errors.
	Err error
}

// New creates a new Error.
//
// It accepts an arbitrary number of arguments of the following types:
//
//     Op
//         The operation during which an error occurred and New was called.
//         Usually the name of the method creating the error. Never the function
//         or method returning an error.
//     Kind
//         The kind of an error. Callers may treat errors differently based
//         on their Kind. If Kind is not specified Unspecified is used as
//         default. Errors with kind Unspecified should always be treated as
//         fatal errors.
//     error
//         An error returned by another function or method.
//     string
//         A string to use as a detailed error message. The string should
//         follow the usual Go conventions for error messages, i.e. it should
//         start with a lower case letter and be relatively short but
//         meaningful.
//
// If more than one argument of the above types is passed, the first passed wins.
// Arguments of other types than the above are silently ignored.
//
// If New is called without any arguments it returns nil. It is rather pointless
// to call New without any arguments. Future versions of New might choose to
// panic instead. It is thus better to not call New without arguments.
func New(args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}
	err := &Error{}
	for _, arg := range args {
		switch v := arg.(type) {
		case Op:
			err.Op = v
		case Kind:
			err.Kind = v
		case error:
			err.Err = v
		case string:
			err.Msg = v
		}
	}
	return err
}

// Wrap returns nil if err is nil. If err is not nil it calls New and returns
// whatever new returns.
func Wrap(err error, args ...interface{}) error {
	if err != nil {
		args = append(args, err)
		return New(args...)
	}
	return nil
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	sb := &strings.Builder{}
	if e.Op != "" {
		sep(sb)
		sb.WriteString(string(e.Op))
	}
	if e.Kind != Unspecified {
		sep(sb)
		sb.WriteString(e.Kind.String())
	}
	if e.Msg != "" {
		sep(sb)
		sb.WriteString(e.Msg)
	}
	if e.Err != nil {
		sep(sb)
		sb.WriteString(e.Err.Error())
	}
	return sb.String()
}

// Trace build a trace of operations that lead to the error.
func (e *Error) Trace() []Op {
	var (
		trace []Op
		cur   *Error = e
	)

	for cur != nil {
		trace = appendTrace(trace, cur)
		if !As(cur.Err, &cur) {
			break
		}
	}

	return trace
}

func appendTrace(trace []Op, err *Error) []Op {
	if err.Op == "" {
		return append(trace, "unknown")
	}
	return append(trace, err.Op)
}

func sep(sb *strings.Builder) {
	if sb.Len() == 0 {
		return
	}
	sb.WriteString(": ")
}

// Unwrap returns the error wrapped by this Error. It returns nil if no error
// is wrapped.
func (e *Error) Unwrap() error {
	return e.Err
}

// Is checks if this error matches target.
//
// A positive match means that target is of type Error and that all non-zero
// fields of target are equal to the respective fields of e.
//
// Is does not compare the Error.Err field. Use the Is function to compare
// all errors in the chain.
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}
	other, ok := target.(*Error)
	if !ok {
		return false
	}
	if other.Op != "" && other.Op != e.Op {
		return false
	}
	if other.Kind != Unspecified && other.Kind != e.Kind {
		return false
	}
	if other.Msg != "" && other.Msg != e.Msg {
		return false
	}
	return true
}

// HasCause returns true if the error err has the error cause in its chain
// of wrapped errors. It returns false otherwise.
//
// Deprecated: Is reports true if any error in the errors chain matches.
func HasCause(err error, cause error) bool {
	if reflect.DeepEqual(err, cause) {
		return true
	}
	var wrapper unwrapper
	if As(err, &wrapper) {
		return HasCause(wrapper.Unwrap(), cause)
	}
	return false
}

type unwrapper interface {
	Unwrap() error
}

// GetKind returns the Kind of the passed error, or Unspecified if the error
// has no Kind or is not an acmeproxy error.
func GetKind(err error) Kind {
	var acpErr *Error

	if err == nil || !As(err, &acpErr) {
		return Unspecified
	}
	if acpErr.Kind == Unspecified {
		return GetKind(acpErr.Err)
	}
	return acpErr.Kind
}

// IsKind checks if the error is of the expected Kind.
func IsKind(err error, kind Kind) bool {
	return GetKind(err) == kind
}

// Match returns true iff err matches the template error tmpl.
//
// Match checks that both tmpl and err are of type *Error. If this is the case
// Match compares every non-zero field of tmpl with the respective field in err.
// If this is not the case it checks the error string of tmpl and err for
// equality.
//
// Match recursively checks tmpl.Err and err.Err if they are set.
//
// Deprecated: with Go 1.13 errors.Is should be used.
func Match(tmpl, err error) bool {
	var (
		tmplErr *Error
		actErr  *Error
	)
	if tmpl == nil || err == nil {
		// true if tmpl and err are nil
		return tmpl == err
	}
	if !As(tmpl, &tmplErr) || !As(err, &actErr) {
		return tmpl.Error() == err.Error()
	}
	if tmplErr.Op != "" && tmplErr.Op != actErr.Op {
		return false
	}
	if tmplErr.Kind != Unspecified && tmplErr.Kind != actErr.Kind {
		return false
	}
	if tmplErr.Msg != "" && tmplErr.Msg != actErr.Msg {
		return false
	}
	if tmplErr.Err != nil {
		return Match(tmplErr.Err, actErr.Err)
	}
	return true
}
