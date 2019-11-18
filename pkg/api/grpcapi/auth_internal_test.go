package grpcapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestTokenAuthCtx(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		parser  TokenParser
		claims  *auth.Claims
		err     error
	}{
		{
			name: "reject context with missing authorization header",
			err:  errors.New(errors.Unauthorized, "missing bearer token"),
		},
		{
			name: "reject context with invalid authorization header",
			headers: map[string]string{
				"authorization": "valid token",
			},
			err: errors.New(errors.Unspecified, "invalid authorization header"),
		},
		{
			name: "reject context with invalid bearer token",
			headers: map[string]string{
				"authorization": "Bearer invalid",
			},
			parser: func(token string) (*auth.Claims, error) {
				if token != "invalid" {
					return nil, errors.New(fmt.Sprintf("expected token value 'invalid'; got %s", token))
				}
				return nil, errors.New(errors.Unauthorized, "invalid bearer token")
			},
			err: errors.New(errors.Unauthorized, "invalid bearer token"),
		},
		{
			name: "accept context with valid bearer token",
			headers: map[string]string{
				"authorization": "Bearer valid",
			},
			parser: func(token string) (*auth.Claims, error) {
				if token != "valid" {
					return nil, errors.New(fmt.Sprintf("expected token value 'valid'; got %s", token))
				}
				return &auth.Claims{
					StandardClaims: jwt.StandardClaims{
						Subject: "jdoe",
					},
				}, nil
			},
			claims: &auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "jdoe",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctxIn := metadata.NewIncomingContext(context.Background(), metadata.New(tt.headers))
			ctx, err := tokenAuthCtx(ctxIn, tt.parser)
			claims, ok := auth.ClaimsFromContext(ctx)
			if !ok && tt.claims != nil {
				t.Error("Expected context to contain claims")
			}
			assert.Equal(t, tt.claims, claims)
			assert.Truef(t, errors.Is(err, tt.err), "expected %v; got %v", tt.err, err)
		})
	}
}

func TestTokenAuthCtx_MissingMetadata(t *testing.T) {
	expectedErr := errors.New(errors.Unspecified, "missing bearer token")
	_, err := tokenAuthCtx(context.Background(), nil)
	assert.Truef(t, errors.Is(err, expectedErr), "expected %v; got %v", expectedErr, err)
}
