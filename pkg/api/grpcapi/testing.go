package grpcapi

import (
	"context"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
)

// ServerTestFixture contains an instance of Server and provides the necessary methods
// to start and stop the server during testing.
type ServerTestFixture struct {
	T      *testing.T
	Server *Server
}

// NewServerTestFixture creates a new ServerTestFixture.
func NewServerTestFixture(t *testing.T) *ServerTestFixture {
	return &ServerTestFixture{
		T:      t,
		Server: &Server{},
	}
}

// Start starts fx.Server in a separate go routine and returns the servers
// address.
func (fx *ServerTestFixture) Start() string {
	addrC := make(chan string)
	go func() {
		err := netutil.ListenAndServe(fx.Server, netutil.NotifyAddr(addrC))
		if err != nil {
			fx.T.Log(err)
		}
		close(addrC)
	}()
	select {
	case addr := <-addrC:
		return addr
	case <-time.After(10 * time.Millisecond):
		fx.T.Fatal("timed out after 10ms")
		return ""
	}
}

// Stop stops the previously started server used by the ServerTestFixture.
func (fx *ServerTestFixture) Stop() {
	if err := fx.Server.Shutdown(context.Background()); err != nil {
		fx.T.Fatal(err)
	}
}
