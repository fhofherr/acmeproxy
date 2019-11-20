package grpcapi_test

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
	"github.com/stretchr/testify/assert"
)

func TestServer_StartCannotBeCalledTwice(t *testing.T) {
	fx := grpcapi.NewTestFixture(t)
	fx.Start()
	defer fx.Stop()

	err := netutil.ListenAndServe(fx.Server)
	assert.Error(t, err)
}

func TestServer_CannotStartServerWithoutTLSConfig(t *testing.T) {
	server := &grpcapi.Server{
		TokenParser: func(string) (*auth.Claims, error) {
			return nil, errors.New(errors.Unauthorized)
		},
	}
	tmpl := errors.New("no tls config provided")
	err := netutil.ListenAndServe(server)
	errors.AssertMatches(t, tmpl, err)
}

func TestServer_CannotStartServerWithoutTokenParser(t *testing.T) {
	server := &grpcapi.Server{
		TLSConfig: &tls.Config{},
	}
	tmpl := errors.New("no token parser provided")
	err := netutil.ListenAndServe(server)
	errors.AssertMatches(t, tmpl, err)
}

func TestServer_CannotShutdownUnstartedServer(t *testing.T) {
	server := &grpcapi.Server{}
	tmpl := errors.New("not started")
	err := server.Shutdown(context.Background())
	errors.AssertMatches(t, tmpl, err)
}

func TestServer_ShutdownClosesServerOnCanceledContext(t *testing.T) {
	fx := grpcapi.NewTestFixture(t)
	fx.Start()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := fx.Server.Shutdown(ctx)
	assert.Error(t, err)
}
