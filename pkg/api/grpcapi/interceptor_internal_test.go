package grpcapi

import (
	"context"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptor_ValidateToken(t *testing.T) {
	tests := []struct {
		name          string
		reqHeaders    map[string]string
		tokenParser   TokenParser
		handlerCalled bool
		code          codes.Code
	}{
		{
			name: "reject invalid token",
			reqHeaders: map[string]string{
				"authorization": "Bearer invalid token",
			},
			tokenParser: func(string) (*auth.Claims, error) {
				return nil, errors.New(errors.Unauthorized)
			},
			code: codes.Unauthenticated,
		},
		{
			name: "accept valid token",
			reqHeaders: map[string]string{
				"authorization": "Bearer valid token",
			},
			tokenParser: func(string) (*auth.Claims, error) {
				return &auth.Claims{}, nil
			},
			handlerCalled: true,
			code:          codes.OK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			interceptor := &unaryServerInterceptor{
				TokenParser: tt.tokenParser,
			}
			handlerCalled := false
			handler := func(context.Context, interface{}) (interface{}, error) {
				handlerCalled = true
				return nil, nil
			}
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(tt.reqHeaders))
			_, err := interceptor.intercept(ctx, nil, nil, handler)
			assert.Equalf(t, tt.handlerCalled, handlerCalled, "handler should have been called: %t", tt.handlerCalled)

			grpcError, ok := status.FromError(err)
			if !ok {
				t.Fatalf("Could not convert %v to grpcError", err)
			}
			assert.Equalf(t, tt.code, grpcError.Code(), "%v did not have code %v", err, tt.code)
		})
	}
}
