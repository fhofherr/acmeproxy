package pb_test

import (
	"fmt"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCErrorConversion(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status *status.Status
	}{
		{
			name: "nil if err is nil",
		},
		{
			name:   "internal error on unrecognized error",
			err:    fmt.Errorf("some error"),
			status: status.New(codes.Internal, "some error"),
		},
		{
			name: "not found error if kind is NotFound",
			err: &errors.Error{
				Kind: errors.NotFound,
				Msg:  "some message",
			},
			status: mustHaveDetails(
				t,
				status.New(codes.NotFound, "some message"),
				&pb.ErrorDetails{
					Kind: int32(errors.NotFound),
					Msg:  "some message",
				}),
		},
		{
			name: "internal error on error of any other kind",
			err: &errors.Error{
				Kind: errors.Unspecified,
				Msg:  "some message",
			},
			status: mustHaveDetails(
				t,
				status.New(codes.Internal, "some message"),
				&pb.ErrorDetails{
					Kind: int32(errors.Unspecified),
					Msg:  "some message",
				}),
		},
		{
			name: "copies all available error fields",
			err: &errors.Error{
				Op:   "some op",
				Kind: errors.NotFound,
				Msg:  "some message",
			},
			status: mustHaveDetails(
				t,
				status.New(codes.NotFound, "some message"),
				&pb.ErrorDetails{
					Op:   "some op",
					Kind: int32(errors.NotFound),
					Msg:  "some message",
				}),
		},
		{
			name: "uses the formatted error on missing Msg field",
			err: &errors.Error{
				Op:   "some op",
				Kind: errors.NotFound,
			},
			status: mustHaveDetails(
				t,
				status.New(
					codes.NotFound,
					fmt.Sprintf("%v", &errors.Error{
						Op:   "some op",
						Kind: errors.NotFound,
					}),
				),
				&pb.ErrorDetails{
					Op:   "some op",
					Kind: int32(errors.NotFound),
				}),
		},
		{
			name: "copies wrapped unrecognized errors",
			err: &errors.Error{
				Msg: "some message",
				Err: fmt.Errorf("wrapped error"),
			},
			status: mustHaveDetails(
				t,
				status.New(codes.Internal, "some message"),
				&pb.ErrorDetails{
					Msg: "some message",
					Err: &pb.ErrorDetails_Plain{
						Plain: fmt.Sprintf("%v", fmt.Errorf("wrapped error")),
					},
				},
			),
		},
		{
			name: "copies wrapped acmeproxy errors",
			err: &errors.Error{
				Msg: "some message",
				Err: &errors.Error{
					Msg: "wrapped error",
				},
			},
			status: mustHaveDetails(
				t,
				status.New(codes.Internal, "some message"),
				&pb.ErrorDetails{
					Msg: "some message",
					Err: &pb.ErrorDetails_Nested{
						Nested: &pb.ErrorDetails{
							Msg: "wrapped error",
						},
					},
				},
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			statusErr := pb.ToGRPCStatusError(tt.err)
			assert.Equal(t, tt.status.Err(), statusErr)
			err := pb.FromGRPCStatusError(statusErr)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestFromGRPCStatusError_NonGRPCError(t *testing.T) {
	err := fmt.Errorf("some error")
	actual := pb.FromGRPCStatusError(err)
	assert.Same(t, err, actual)
}

func mustHaveDetails(t *testing.T, st *status.Status, details ...proto.Message) *status.Status {
	var err error

	st, err = st.WithDetails(details...)
	if err != nil {
		t.Fatal(err)
	}
	return st
}
