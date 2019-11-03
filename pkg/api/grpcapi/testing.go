package grpcapi

import (
	"context"
	"crypto/tls"
	"path/filepath"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
)

// TestFixture contains an instance of Server and provides the necessary methods
// to start and stop the server during testing.
type TestFixture struct {
	T         *testing.T
	Server    *Server
	TLSConfig *tls.Config
}

// NewTestFixture creates a new TestFixture.
func NewTestFixture(t *testing.T) *TestFixture {
	keyFile := filepath.Join("testdata", "key.pem")
	certFile := filepath.Join("testdata", "cert.pem")
	if *testsupport.FlagUpdate {
		certutil.CreateOpenSSLPrivateKey(t, certutil.RSA2048, keyFile, true)
		certutil.CreateOpenSSLSelfSignedCertificate(t, "localhost", keyFile, certFile, true)
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	tlsConfig := &tls.Config{
		// This is ok for testing. Do not use this for production code!
		// nolint: gosec
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}
	return &TestFixture{
		T:         t,
		TLSConfig: tlsConfig,
		Server: &Server{
			TLSConfig: tlsConfig,
		},
	}
}

// Start starts fx.Server in a separate go routine and returns the servers
// address.
func (fx *TestFixture) Start() string {
	addrC := make(chan string)
	go func() {
		err := netutil.ListenAndServe(fx.Server, netutil.NotifyAddr(addrC))
		if err != nil {
			fx.T.Log(err)
		}
	}()
	select {
	case addr := <-addrC:
		return addr
	case <-time.After(10 * time.Millisecond):
		fx.T.Fatal("timed out after 10ms")
		return ""
	}
}

// Stop stops the previously started server used by the TestFixture.
func (fx *TestFixture) Stop() {
	if err := fx.Server.Shutdown(context.Background()); err != nil {
		fx.T.Fatal(err)
	}
}

// NewClient creates a new GRPCApi client connecting to the server contained in
// this test fixture.
func (fx *TestFixture) NewClient(addr string) *Client {
	client, err := NewClient(addr, fx.TLSConfig)
	if err != nil {
		fx.T.Fatal(err)
	}
	return client
}
