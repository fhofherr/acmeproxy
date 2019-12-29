package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi"
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
//
// Server starts multiple go routines but is not itself safe for concurrent
// access.
type Server struct {
	ACMEDirectoryURL string
	HTTPAPIAddr      string
	GRPCAPIAddr      string
	DataDir          string

	// The fully quallyfied domain name under which acmeproxy is accessible.
	// Defaults to localhost.localdomain. Outside of tests this is most likely
	// not what you want.
	AcmeproxyFQDN string

	// ID of the default account. If this is empty Server creates a new random
	// UUID during every start. This in turn leads to a new account being
	// created on the ACME CA.
	DefaultUserID uuid.UUID

	// Email used to create a new ACME account. If this is empty an account
	// without notification address is created.
	DefaultAccountEmail string

	Logger log.Logger

	httpAPIServer *httpapi.Server
	grpcAPIServer *grpcapi.Server
	acmeAgent     *acme.Agent
	boltDB        *db.Bolt
	once          sync.Once
	initErr       error

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
		if s.initErr = s.registerAcmeproxyDomain(); s.initErr != nil {
			return
		}
		if s.initErr = s.startGRPCAPI(); s.initErr != nil {
			return
		}
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
	case addr := <-addrC:
		s.HTTPAPIAddr = addr
		log.Log(s.Logger, "level", "info", "msg", fmt.Sprintf("HTTP API is listening on: %s", addr))
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New(op, "http API startup timed out")
	}
}

func (s *Server) registerAcmeproxyDomain() error {
	const op errors.Op = "server/server.registerAcmeproxyDomain"

	if len(s.DefaultUserID) == 0 {
		var err error

		s.DefaultUserID, err = uuid.NewRandom()
		if err != nil {
			return errors.New(op, "create random userID", err)
		}
	}
	if s.AcmeproxyFQDN == "" {
		s.AcmeproxyFQDN = "localhost.localdomain"
	}
	if err := s.acmeAgent.RegisterUser(s.DefaultUserID, s.DefaultAccountEmail); err != nil {
		return errors.New(op, "register default user", err)
	}
	if err := s.acmeAgent.RegisterDomain(s.DefaultUserID, s.AcmeproxyFQDN); err != nil {
		return errors.New(op, fmt.Sprintf("register acmeproxy domain: %s", s.AcmeproxyFQDN), err)
	}
	return nil
}

func (s *Server) startGRPCAPI() error {
	const op errors.Op = "api/server.startGRPCAPI"
	var (
		cert    tls.Certificate
		certBuf bytes.Buffer
		keyBuf  bytes.Buffer
		err     error
	)

	if err = s.acmeAgent.WriteCertificate(s.DefaultUserID, s.AcmeproxyFQDN, &certBuf); err != nil {
		return errors.New(op, err)
	}
	if err = s.acmeAgent.WritePrivateKey(s.DefaultUserID, s.AcmeproxyFQDN, &keyBuf); err != nil {
		return errors.New(op, err)
	}
	if cert, err = tls.X509KeyPair(certBuf.Bytes(), keyBuf.Bytes()); err != nil {
		return errors.New(op, err)
	}

	addrC := make(chan string)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	s.grpcAPIServer = &grpcapi.Server{
		TLSConfig: tlsConfig,
		TokenParser: func(token string) (*auth.Claims, error) {
			// TODO make public key for JWTs configurable
			// return auth.ParseToken(token, alg, key)
			return nil, errors.New(errors.Unauthorized)
		},
		UserRegisterer: s.acmeAgent,
		Logger:         s.Logger,
	}
	go errors.LogFunc(s.Logger, func() error {
		return netutil.ListenAndServe(s.grpcAPIServer, netutil.WithAddr(s.GRPCAPIAddr), netutil.NotifyAddr(addrC))
	})

	select {
	case addr := <-addrC:
		s.GRPCAPIAddr = addr
		log.Log(s.Logger, "level", "info", "msg", fmt.Sprintf("GRPC API is listening on: %s", addr))
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New(op, "http API startup timed out")
	}
}

// Shutdown performs a graceful shutdown of the Server. Once the server has
// been shut down, it cannot be started again.
func (s *Server) Shutdown(ctx context.Context) error {
	const op errors.Op = "server/server.Shutdown"

	if atomic.LoadUint32(&s.started) == 0 {
		return errors.New(op, "not started")
	}

	var errcol errors.Collection
	errcol = errors.Append(errcol, s.grpcAPIServer.Shutdown(ctx), op)
	errcol = errors.Append(errcol, s.httpAPIServer.Shutdown(ctx), op)
	errcol = errors.Append(errcol, s.boltDB.Close(), op)
	return errcol.ErrorOrNil()
}
