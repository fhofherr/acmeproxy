package errors

import (
	"strings"
)

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
)

func (k Kind) String() string {
	switch k {
	case Unspecified:
		return ""
	case NotFound:
		return "not found"
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

func sep(sb *strings.Builder) {
	if sb.Len() == 0 {
		return
	}
	sb.WriteString(": ")
}
