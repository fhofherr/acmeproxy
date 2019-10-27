package grpcapi

import (
	"context"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type adminServer struct {
}

func (s *adminServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	st, _ := status.
		New(codes.NotFound, "not found").
		WithDetails(&pb.Error{Op: "some op"})
	return nil, st.Err()
}
