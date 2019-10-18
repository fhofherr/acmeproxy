package server

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/httpapi"
	"github.com/fhofherr/golf/log"
	"github.com/google/uuid"
)

// Server is acmeproxy's public server.
//
// Server runs the ACME Agent responsible of obtaining certificates and storing
// them for later retrieval. Users connect to the server through its public
// API to retrieve their certificates.
//
// The zero value of Server represents a valid instance. Server may start
// a multitude of Go routines.
type Server struct {
	ACMEDirectoryURL string
	HTTPAPIAddr      string
	DataDir          string
	Logger           log.Logger
	httpAPIServer    *httpServer
	acmeAgent        *acme.Agent
	boltDB           *db.Bolt

	// Accessed atomically. A non-zero value means the server is currently
	// starting or has already been started. It cannot be started again.
	started uint32
}

// Start starts the Server without blocking.
//
// A Server can be started only once. Re-starting an already started server
// -- even if it has been stopped in the meantime by calling Shutdown -- leads
// to an error.
func (s *Server) Start() error {
	const op errors.Op = "server/server.Start"

	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New(op, "already started")
	}
	s.initialize()

	if err := s.boltDB.Open(); err != nil {
		return errors.New(op, err)
	}
	go func() {
		errors.Log(s.Logger, s.httpAPIServer.Start())
	}()

	if err := s.registerAcmeproxyDomain(); err != nil {
		return errors.New(op, err)
	}
	return nil
}

// isStarted returns true if an attempt to start the server has been made. A
// true return value does not indicate the server is actually running. Start
// may have failed with an error, or Shutdown was called in the meantime.
func (s *Server) isStarted() bool {
	return atomic.LoadUint32(&s.started) > 0
}

// Shutdown performs a graceful shutdown of the Server. Once the server has
// been shut down, it cannot be started again.
func (s *Server) Shutdown(ctx context.Context) error {
	const op errors.Op = "server/server.Shutdown"

	if !s.isStarted() {
		return errors.New(op, "not started")
	}

	var errcol errors.Collection
	errcol = errors.Append(errcol, s.httpAPIServer.Shutdown(ctx), op)
	errcol = errors.Append(errcol, s.boltDB.Close(), op)
	return errcol.ErrorOrNil()
}

// initialize initializes the un-exported fields of Server. It must not
// be called more than once.
func (s *Server) initialize() {
	s.boltDB = &db.Bolt{
		FilePath: filepath.Join(s.DataDir, "acmeproxy.db"),
		FileMode: 0600,
	}
	acmeclient.InitializeLego(s.Logger)
	acmeClient := &acmeclient.Client{
		DirectoryURL: s.ACMEDirectoryURL,
	}
	s.acmeAgent = &acme.Agent{
		Domains:      s.boltDB.DomainRepository(),
		Users:        s.boltDB.UserRepository(),
		Certificates: acmeClient,
		ACMEAccounts: acmeClient,
	}
	s.httpAPIServer = &httpServer{
		Addr:   s.HTTPAPIAddr,
		Logger: s.Logger,
		Handler: httpapi.NewRouter(httpapi.Config{
			Solver: &acmeClient.HTTP01Solver,
		}),
	}
}
func (s *Server) registerAcmeproxyDomain() error {
	const op errors.Op = "server/server.registerAcmeproxyDomain"

	tmpUserID := uuid.Must(uuid.NewRandom())
	if err := s.acmeAgent.RegisterUser(tmpUserID, ""); err != nil {
		return errors.New(op, "register default user", err)
	}
	if err := s.acmeAgent.RegisterDomain(tmpUserID, "www.example.com"); err != nil {
		return errors.New(op, fmt.Sprintf("register acmeproxy domain: %s", "www.example.com"), err)
	}
	return nil
}
