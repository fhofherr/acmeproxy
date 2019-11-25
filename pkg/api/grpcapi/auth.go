package grpcapi

import (
	"context"
	"strings"

	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"google.golang.org/grpc/metadata"
)

// TokenParser is a function that parses and validates the signature of the
// passed token string.
//
// If the signature is valid it returns the claims contained in the token.
// Otherwise it returns an error.
type TokenParser func(string) (*auth.Claims, error)

func tokenAuthCtx(ctx context.Context, parse TokenParser) (context.Context, error) {
	const op errors.Op = "grpcapi/tokenAuthCtx"

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, errors.New(op, errors.Unauthorized, "missing bearer token")
	}
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return ctx, errors.New(op, errors.Unauthorized, "missing bearer token")
	}
	token := authHeader[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return ctx, errors.New(op, errors.Unauthorized, "invalid authorization header")
	}
	token = strings.TrimPrefix(token, "Bearer ")
	claims, err := parse(token)
	if err != nil {
		return ctx, errors.New(op, err)
	}
	return auth.AddClaimsToContext(ctx, claims), nil
}
