package grpcapi_test

import (
	"context"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUser(t *testing.T) {
	fx := grpcapi.NewServerTestFixture(t)
	addr := fx.Start()

	client, err := grpcapi.NewClient(addr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.RegisterUser(context.Background(), "john.doe@example.com")
	assert.True(t, errors.Is(err, errors.New(errors.NotFound)))
}
