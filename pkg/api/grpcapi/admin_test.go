package grpcapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterUser_Authorization(t *testing.T) {
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

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			email := "acmeproxy.user@example.com"

			fx.MockUserRegisterer.
				On("RegisterUser", mock.AnythingOfType("uuid.UUID"), email).
				Return(nil)

			userID, err := client.RegisterUser(ctx, email)
			if err == nil {
				fx.MockUserRegisterer.AssertCalled(
					t, "RegisterUser", mock.AnythingOfType("uuid.UUID"), email)
				assert.NotEqual(t, uuid.UUID{}, userID)
			}
			errors.AssertMatches(t, tt.err, err)
		})
	}
}
