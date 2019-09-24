package server

import (
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
)

// TestFixture wraps everything required to test Server.
//
// Use NewTestFixture to create a working instance.
type TestFixture struct {
	DataDir  string
	Pebble   *testsupport.Pebble
	Server   *Server
	rmTmpDir func()
}

// NewTestFixture creates a new test fixture ready for use.
func NewTestFixture(t *testing.T) *TestFixture {
	tmpDir, rmTmpDir := testsupport.CreateTmpDir(t)
	pebble := testsupport.NewPebble(t)
	dataDir := filepath.Join(tmpDir, "data")
	server := &Server{
		DataDir: dataDir,
	}
	return &TestFixture{
		Server:   server,
		DataDir:  dataDir,
		Pebble:   pebble,
		rmTmpDir: rmTmpDir,
	}
}

// Close destroys all data structures and directories associated with this
// TestFixture.
func (fx *TestFixture) Close() error {
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
