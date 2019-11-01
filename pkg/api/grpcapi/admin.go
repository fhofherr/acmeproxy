package grpcapi

import (
	"context"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type adminServer struct {
}

func (s *adminServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	const op errors.Op = "grpcapi/adminServer.RegisterUser"

	err := errors.New(op, errors.NotFound)
	return nil, pb.ToGRPCStatusError(err)
}

type adminClient struct {
	Client pb.AdminClient
}

func (c *adminClient) RegisterUser(ctx context.Context, email string) (uuid.UUID, error) {
	const op errors.Op = "grpcapi/adminClient.RegisterUser"

	req := &pb.RegisterUserRequest{
		Email: email,
	}
	_, err := c.Client.RegisterUser(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			// TODO add proper translation from/to errors.Kind to codes.Code
			kind := errors.Unspecified
			if st.Code() == codes.NotFound {
				kind = errors.NotFound
			}
			// TODO marshal details back into errors.Error
			return uuid.UUID{}, errors.New(op, kind, st.Details())
		}
		// TODO add a proper error message
		return uuid.UUID{}, errors.New(op, err)
	}
	return uuid.UUID{}, nil
}
