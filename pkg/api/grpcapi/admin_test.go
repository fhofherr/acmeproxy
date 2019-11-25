package grpcapi_test

import (
	"context"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
	"github.com/fhofherr/acmeproxy/pkg/errors"
)

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		claims *auth.Claims
		err    error
	}{
		{
			name:  "invalid token",
			token: "invalid",
			err:   errors.New(errors.Unauthorized),
		},
		{
			name:  "valid token but no roles",
			token: "valid",
			claims: &auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "jdoe@example.com",
				},
			},
			err: errors.New(errors.Unauthorized),
		},
		{
			name:  "valid token but not an admin",
			token: "valid",
			claims: &auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "jdoe@example.com",
				},
				Roles: []auth.Role{auth.Role("role1"), auth.Role("role2")},
			},
			err: errors.New(errors.Unauthorized),
		},
		{
			name:  "valid token and an admin",
			token: "valid",
			claims: &auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "jdoe@example.com",
				},
				Roles: []auth.Role{auth.Admin},
			},
			err: errors.New(errors.NotFound),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fx := grpcapi.NewTestFixture(t)
			fx.Token = "valid"
			fx.Claims = tt.claims
			addr := fx.Start()
			defer fx.Stop()
			client := fx.NewClient(addr, tt.token)
			_, err := client.RegisterUser(context.Background(), "acmeproxy.user@example.com")
			errors.AssertMatches(t, tt.err, err)
		})
	}
}
