package grpcapi

import (
	"crypto/tls"
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client is used to connect to the Server using grpc.
type Client struct {
	adminClient
	conn *grpc.ClientConn
}

// NewClient creates a new Client connecting to the server listening on addr.
func NewClient(addr string, tlsConfig *tls.Config) (*Client, error) {
	const op errors.Op = "grpcapi/NewClient"

	creds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
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
