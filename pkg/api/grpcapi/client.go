package grpcapi

import (
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"google.golang.org/grpc"
)

// Client is used to connect to the Server using grpc.
type Client struct {
	adminClient
	conn *grpc.ClientConn
}

// NewClient creates a new Client connecting to the server listening on addr.
func NewClient(addr string) (*Client, error) {
	const op errors.Op = "grpcapi/NewClient"
	// TODO add proper transport security
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return &Client{}, errors.New(op, fmt.Sprintf("dial: %s", addr), err)
	}
	return &Client{
		adminClient: adminClient{
			Client: pb.NewAdminClient(conn),
		},
		conn: conn,
	}, nil
}
