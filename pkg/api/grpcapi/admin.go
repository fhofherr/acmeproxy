package grpcapi

import (
	"context"

	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
)

// UserRegisterer wraps the register user method.
//
// RegisterUser allows to create a new user with the passed userID and email.
// Implementations may treat the email as optional and thus may accept an
// empty string for email. If an user with the passed userID already exists
// implementations should do nothing and especially the must not return an
// error.
type UserRegisterer interface {
	RegisterUser(userID uuid.UUID, email string) error
}

type adminServer struct {
	UserRegisterer UserRegisterer
}

func (s *adminServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	const op errors.Op = "grpcapi/adminServer.RegisterUser"

	if err := auth.CheckRoles(ctx, auth.Admin); err != nil {
		return nil, pb.ToGRPCStatusError(err)
	}

	userID, err := uuid.NewRandom()
	if err != nil {
		err = errors.New(op, "new random UUID", err)
		return nil, pb.ToGRPCStatusError(err)
	}
	// TODO find a way to re-use the same code in db/internal/dbrecords
	idBytes, err := userID.MarshalBinary()
	if err != nil {
		err = errors.New(op, "binary marshal UUID")
		return nil, pb.ToGRPCStatusError(err)
	}
	email := req.GetEmail()

	// TODO test RegisterUser returns error
	if err := s.UserRegisterer.RegisterUser(userID, email); err != nil {
		return nil, pb.ToGRPCStatusError(err)
	}
	// TODO why did I wrap this into a RegisterUserResponse? Refactor?
	return &pb.RegisterUserResponse{
		User: &pb.User{
			Id:    idBytes,
			Email: email,
		},
	}, nil
}

type adminClient struct {
	Client pb.AdminClient
}

func (c *adminClient) RegisterUser(ctx context.Context, email string) (uuid.UUID, error) {
	const op errors.Op = "grpcapi/adminClient.RegisterUser"

	req := &pb.RegisterUserRequest{
		Email: email,
	}
	res, err := c.Client.RegisterUser(ctx, req)
	if err != nil {
		err = pb.FromGRPCStatusError(err)
		// TODO add a proper error message
		return uuid.UUID{}, errors.New(op, err)
	}
	idBytes := res.GetUser().GetId()
	userID, err := uuid.FromBytes(idBytes)
	if err != nil {
		return uuid.UUID{}, errors.New(op, "unmarshal userID", err)
	}
	return userID, nil
}
