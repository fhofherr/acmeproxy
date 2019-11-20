package grpcapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
)

// TestFixture contains an instance of Server and provides the necessary methods
// to start and stop the server during testing.
type TestFixture struct {
	T         *testing.T
	Server    *Server
	TLSConfig *tls.Config
	Token     string
	Claims    *auth.Claims
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
	fx := &TestFixture{
		T:         t,
		TLSConfig: tlsConfig,
	}
	server := &Server{
		TLSConfig:   tlsConfig,
		TokenParser: fx.parseToken,
	}

	fx.Server = server
	return fx
}

func (fx *TestFixture) parseToken(token string) (*auth.Claims, error) {
	const op errors.Op = "grpcapi/testFixture.parseToken"

	if fx.Token == "" {
		msg := "fx.Token is empty; rejecting all authentication attempts"
		return nil, errors.New(op, errors.Unauthorized, msg)
	}
	if fx.Token != token {
		msg := fmt.Sprintf("expected token '%s' got '%s'", fx.Token, token)
		return nil, errors.New(op, errors.Unauthorized, msg)
	}
	return fx.Claims, nil
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
func (fx *TestFixture) NewClient(addr, token string) *Client {
	authToken := &AuthToken{Token: token}
	client, err := NewClient(addr, authToken, fx.TLSConfig)
	if err != nil {
		fx.T.Fatal(err)
	}
	return client
}
