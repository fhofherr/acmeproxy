package pb

import (
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCStatusError translates err to a status as defined by
// google.golang.org/grpc/status and returns the result of status/Status.Err().
//
// If err is a *errors.Error ToGRPCStatusError uses the following operations
// to transform err:
//
//     * The Kind of err is used to determine the GRPC status code.
//
//     * If the Msg field is set the value of Msg is used as the status message.
//       If Msg is empty the result of fmt.Sprint("%v", err) is used as status
//       message.
//
//     * err itself is transformed to an *ErrorDetails and attached to the
//       status as details.
//
// If err is an arbitrary error the following operations are used instead:
//
//     * The status code is set to google.golang.org/grpc/codes.Internal.
//
//     * The result of fmt.Sprintf("%v", err) is used as the error message.
//
//     * No details are appended to the Status.
//
// If err is nil ToGRPCStatusError returns nil.
func ToGRPCStatusError(err error) error {
	var acpErr *errors.Error

	if err == nil {
		return nil
	}
	if !errors.As(err, &acpErr) {
		return status.New(codes.Internal, fmt.Sprintf("%v", err)).Err()
	}
	code := codeFromErr(acpErr)
	msg := acpErr.Msg
	if msg == "" {
		msg = fmt.Sprintf("%v", acpErr)
	}
	details := errToDetails(acpErr)
	st, err := status.New(code, msg).WithDetails(details)
	if err != nil {
		// This should never happen as we know the details passed can be
		// marshaled. Nevertheless we want to make sure, we get an error
		// if we are wrong and this occurs.
		msg := fmt.Sprintf("grpcapi: pb/ToGRPCStatusError: %v", err)
		return status.New(codes.Internal, msg).Err()
	}
	return st.Err()
}

func codeFromErr(err *errors.Error) codes.Code {
	switch err.Kind {
	case errors.NotFound:
		return codes.NotFound
	default:
		return codes.Internal
	}
}

func errToDetails(err *errors.Error) *ErrorDetails {
	details := &ErrorDetails{
		Op:   string(err.Op),
		Kind: int32(err.Kind),
		Msg:  err.Msg,
	}
	if err.Err == nil {
		return details
	}

	var acpErr *errors.Error
	if errors.As(err.Err, &acpErr) {
		details.Err = &ErrorDetails_Nested{
			Nested: errToDetails(acpErr),
		}
	} else {
		details.Err = &ErrorDetails_Plain{
			Plain: fmt.Sprintf("%v", err.Err),
		}
	}
	return details
}
