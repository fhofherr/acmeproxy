package api

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/api/httpapi"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
	"github.com/fhofherr/golf/log"
	"github.com/google/uuid"
)

// Server is acmeproxy's public server.
//
// Server runs the ACME Agent responsible of obtaining certificates and storing
// them for later retrieval. Users connect to the server through its public
// API to retrieve their certificates.
type Server struct {
	ACMEDirectoryURL string
	HTTPAPIAddr      string
	DataDir          string
	Logger           log.Logger
	httpAPIServer    *httpapi.Server
	acmeAgent        *acme.Agent
	boltDB           *db.Bolt

	once    sync.Once
	initErr error

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
	const op errors.Op = "api/server.Start"

	if err := s.startOnce(); err != nil {
		return errors.New(op, err)
	}

	return nil
}

func (s *Server) startOnce() error {
	const op errors.Op = "api/server.startOnce"

	s.once.Do(func() {
		acmeclient.InitializeLego(s.Logger)
		if s.initErr = s.initDB(); s.initErr != nil {
			return
		}
		acmeClient := &acmeclient.Client{
			DirectoryURL: s.ACMEDirectoryURL,
		}
		s.acmeAgent = &acme.Agent{
			Domains:      s.boltDB.DomainRepository(),
			Users:        s.boltDB.UserRepository(),
			Certificates: acmeClient,
			ACMEAccounts: acmeClient,
		}
		s.httpAPIServer = &httpapi.Server{
			Solver: &acmeClient.HTTP01Solver,
		}
		if s.initErr = s.startHTTPAPI(); s.initErr != nil {
			return
		}
		// TODO make acmeproxy domain configurable (#41)
		if s.initErr = s.registerAcmeproxyDomain(); s.initErr != nil {
			return
		}
		// TODO obtain certificates for gRPC API
		// TODO make public key for JWTs configurable (#41)
		// TODO start gRPC API
	})

	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New(op, "already started")
	}
	return s.initErr
}

func (s *Server) initDB() error {
	const op errors.Op = "api/Server.initDB"

	s.boltDB = &db.Bolt{
		FilePath: filepath.Join(s.DataDir, "acmeproxy.db"),
		FileMode: 0600,
	}
	err := s.boltDB.Open()
	return errors.Wrap(err, op)
}

func (s *Server) startHTTPAPI() error {
	const op errors.Op = "api/server.startHTTPAPI"

	addrC := make(chan string)
	go errors.LogFunc(s.Logger, func() error {
		err := netutil.ListenAndServe(s.httpAPIServer, netutil.WithAddr(s.HTTPAPIAddr), netutil.NotifyAddr(addrC))
		return errors.Wrap(err, op)
	})
	select {
	case <-addrC:
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New(op, "http API startup timed out")
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

// Shutdown performs a graceful shutdown of the Server. Once the server has
// been shut down, it cannot be started again.
func (s *Server) Shutdown(ctx context.Context) error {
	const op errors.Op = "server/server.Shutdown"

	if atomic.LoadUint32(&s.started) == 0 {
		return errors.New(op, "not started")
	}

	var errcol errors.Collection
	errcol = errors.Append(errcol, s.httpAPIServer.Shutdown(ctx), op)
	errcol = errors.Append(errcol, s.boltDB.Close(), op)
	return errcol.ErrorOrNil()
}
