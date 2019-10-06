package server

import (
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
)

const (
	// PebbleConfigJSON contains the configration of Pebble.
	PebbleConfigJSON = "testdata/pebble-config.json"
	// DNSPort is the port pebble uses for DNS queries.
	DNSPort = "9053"
)

// TestFixture wraps everything required to test Server.
//
// Use NewTestFixture to create a working instance.
type TestFixture struct {
	DataDir  string
	Pebble   *testsupport.Pebble
	Server   *Server
	t        *testing.T
	rmTmpDir func()
}

// NewTestFixture creates a new test fixture ready for use.
func NewTestFixture(t *testing.T) *TestFixture {
	tmpDir, rmTmpDir := testsupport.CreateTmpDir(t)
	pebble := testsupport.NewPebble(t, PebbleConfigJSON, DNSPort)
	dataDir := filepath.Join(tmpDir, "data")
	server := &Server{
		DataDir: dataDir,
	}
	return &TestFixture{
		Server:   server,
		DataDir:  dataDir,
		Pebble:   pebble,
		t:        t,
		rmTmpDir: rmTmpDir,
	}
}

// Close destroys all data structures and directories associated with this
// TestFixture.
func (fx *TestFixture) Close() error {
	fx.Pebble.Stop(fx.t)
	fx.rmTmpDir()
	return nil
}

// MustStartServer calls fx.Server.Start and fails the test if Start returns
// an error.
func (fx *TestFixture) MustStartServer(t *testing.T) {
	err := fx.Server.Start()
	if err != nil {
		t.Fatal(err)
	}
}
