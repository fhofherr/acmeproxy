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

func TestToGRPCStatusError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected *status.Status
	}{
		{
			name: "nil if err is nil",
		},
		{
			name:     "internal error on unrecognized error",
			err:      fmt.Errorf("some error"),
			expected: status.New(codes.Internal, "some error"),
		},
		{
			name: "not found error if kind is NotFound",
			err: &errors.Error{
				Kind: errors.NotFound,
				Msg:  "some message",
			},
			expected: mustHaveDetails(
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
			expected: mustHaveDetails(
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
			expected: mustHaveDetails(
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
			expected: mustHaveDetails(
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
			expected: mustHaveDetails(
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
			expected: mustHaveDetails(
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
			actual, ok := status.FromError(pb.ToGRPCStatusError(tt.err))
			if !ok {
				t.Fatalf("Failed to covert actual error to status")
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func mustHaveDetails(t *testing.T, st *status.Status, details ...proto.Message) *status.Status {
	var err error

	st, err = st.WithDetails(details...)
	if err != nil {
		t.Fatal(err)
	}
	return st
}
