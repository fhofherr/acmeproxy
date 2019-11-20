package grpcapi_test

import (
	"context"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
	"github.com/fhofherr/acmeproxy/pkg/errors"
)

func TestRegisterUser(t *testing.T) {
	fx := grpcapi.NewTestFixture(t)
	fx.Token = "valid"
	addr := fx.Start()
	client := fx.NewClient(addr, fx.Token)

	_, err := client.RegisterUser(context.Background(), "john.doe@example.com")
	errors.AssertMatches(t, errors.New(errors.NotFound), err)
}
