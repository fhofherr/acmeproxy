package grpcapi

import (
	"context"

	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
)

type adminServer struct {
}

func (s *adminServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	const op errors.Op = "grpcapi/adminServer.RegisterUser"

	if err := auth.CheckRoles(ctx, auth.Admin); err != nil {
		return nil, pb.ToGRPCStatusError(err)
	}

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
		err = pb.FromGRPCStatusError(err)
		// TODO add a proper error message
		return uuid.UUID{}, errors.New(op, err)
	}
	return uuid.UUID{}, nil
}
