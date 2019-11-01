package grpcapi_test

import (
	"context"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
	"github.com/stretchr/testify/assert"
)

func TestServer_StartCannotBeCalledTwice(t *testing.T) {
	fx := grpcapi.NewServerTestFixture(t)
	fx.Start()
	defer fx.Stop()

	err := netutil.ListenAndServe(fx.Server)
	assert.Error(t, err)
}

func TestServer_CannotShutdownUnstartedServer(t *testing.T) {
	server := &grpcapi.Server{}
	assert.Error(t, server.Shutdown(context.Background()))
}

func TestServer_ShutdownClosesServerOnCanceledContext(t *testing.T) {
	fx := grpcapi.NewServerTestFixture(t)
	fx.Start()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := fx.Server.Shutdown(ctx)
	assert.Error(t, err)
}
