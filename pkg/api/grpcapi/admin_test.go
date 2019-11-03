package grpcapi_test

import (
	"context"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUser(t *testing.T) {
	fx := grpcapi.NewTestFixture(t)
	addr := fx.Start()
	client := fx.NewClient(addr)

	_, err := client.RegisterUser(context.Background(), "john.doe@example.com")
	t.Log(err)
	assert.True(t, errors.Is(err, errors.New(errors.NotFound)))
}
