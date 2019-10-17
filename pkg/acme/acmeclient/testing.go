package acmeclient

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
)

const (
	// PebbleConfigJSON contains the configration of Pebble.
	PebbleConfigJSON = "testdata/pebble-config.json"
	// DNSPort is the port pebble uses for DNS queries.
	DNSPort = "8053"
)

// TestFixture wraps a Client suitable for testing.
type TestFixture struct {
	Pebble *testsupport.Pebble
	Client *Client
}

// NewTestFixture creates a new test fixture.
func NewTestFixture(t *testing.T) (TestFixture, func()) {
	pebble := testsupport.NewPebble(t, PebbleConfigJSON, DNSPort)
	pebble.Start(t)

	resetCACerts := testsupport.SetLegoCACertificates(t, pebble.TestCert)
	client := &Client{
		DirectoryURL: pebble.DirectoryURL(),
	}
	server := NewChallengeServer(t, &client.HTTP01Solver, pebble.HTTPPort())
	fixture := TestFixture{
		Pebble: pebble,
		Client: client,
	}
	return fixture, func() {
		server.Close()
		resetCACerts()
		pebble.Stop(t)
	}
}

// NewChallengeServer creates an httptest.Server which uses the handler to
// serve HTTP01 challenges.
func NewChallengeServer(t *testing.T, solver *HTTP01Solver, port int) *httptest.Server {
	server := httptest.NewUnstartedServer(solver.Handler(func(req *http.Request) map[string]string {
		portSuffix := fmt.Sprintf(":%d", port)
		domain := strings.Replace(req.Host, portSuffix, "", -1)
		pathParts := strings.Split(req.URL.Path, "/")
		if len(pathParts) == 0 {
			t.Fatalf("Could not obtain token from url: %v", req.URL)
		}
		token := pathParts[len(pathParts)-1]
		return map[string]string{
			"domain": domain,
			"token":  token,
		}
	}))
	address := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	server.Listener = listener
	server.Start()
	return server
}
