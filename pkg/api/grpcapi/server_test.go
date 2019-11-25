package grpcapi_test

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

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

func TestServer_CannotStartWithoutRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		server *grpcapi.Server
		err    error
	}{
		{
			name: "cannot start without tls config",
			server: &grpcapi.Server{
				TokenParser: alwaysUnauthorized,
			},
			err: errors.New("no tls config provided"),
		},
		{
			name: "cannot start server without token parser",
			server: &grpcapi.Server{
				TLSConfig: &tls.Config{},
			},
			err: errors.New("no token parser provided"),
		},
		{
			name: "cannot start server without user registerer",
			server: &grpcapi.Server{
				TLSConfig:   &tls.Config{},
				TokenParser: alwaysUnauthorized,
			},
			err: errors.New("no user registerer provided"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errc := make(chan error)
			go func() {
				errc <- netutil.ListenAndServe(tt.server)
			}()

			select {
			case err := <-errc:
				errors.AssertMatches(t, tt.err, err)
			case <-time.After(10 * time.Millisecond):
				t.Error("server should not have started")
				tt.server.Shutdown(context.Background()) // nolint: errcheck
			}
		})
	}
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

func alwaysUnauthorized(string) (*auth.Claims, error) {
	return nil, errors.New(errors.Unauthorized)
}
