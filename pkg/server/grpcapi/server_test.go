package grpcapi_test

import (
	"context"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
	"github.com/fhofherr/acmeproxy/pkg/server/grpcapi"
	"github.com/stretchr/testify/assert"
)

func TestServer_StartCannotBeCalledTwice(t *testing.T) {
	addrC := make(chan string)
	server := &grpcapi.Server{}
	go netutil.ListenAndServe(server, netutil.NotifyAddr(addrC)) // nolint: errcheck
	defer server.Shutdown(context.Background())                  // nolint: errcheck

	// We don't care for the addr here. Calling GetAddr merely ensures the
	// server is started and ready.
	_ = netutil.GetAddr(t, addrC)
	err := netutil.ListenAndServe(server)
	assert.Error(t, err)
}

func TestServer_CannotShutdownUnstartedServer(t *testing.T) {
	server := &grpcapi.Server{}
	assert.Error(t, server.Shutdown(context.Background()))
}

func TestServer_ShutdownClosesServerOnCanceledContext(t *testing.T) {
	addrC := make(chan string)
	server := &grpcapi.Server{}
	go netutil.ListenAndServe(server, netutil.NotifyAddr(addrC)) // nolint: errcheck
	_ = netutil.GetAddr(t, addrC)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := server.Shutdown(ctx)
	assert.Error(t, err)
}
