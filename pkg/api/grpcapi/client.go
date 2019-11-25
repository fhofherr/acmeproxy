package grpcapi

import (
	"context"
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
func NewClient(addr string, token *AuthToken, tlsConfig *tls.Config) (*Client, error) {
	const op errors.Op = "grpcapi/NewClient"

	transportCreds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(token),
	)
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

// AuthToken represents a fixed authorization token used to authenticate
// the client with the server.
type AuthToken struct {
	Token string
}

// GetRequestMetadata returns the current request metdadata.
//
// See https://godoc.org/google.golang.org/grpc/credentials#PerRPCCredentials
func (t *AuthToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.Token,
	}, nil
}

// RequireTransportSecurity indicates that this authentication method requires
// a secure connection.
//
// See https://godoc.org/google.golang.org/grpc/credentials#PerRPCCredentials
func (t *AuthToken) RequireTransportSecurity() bool {
	return true
}
