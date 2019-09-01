package server

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/fhofherr/acmeproxy/pkg/httpapi"
	"github.com/fhofherr/golf/log"
	"github.com/google/uuid"
)

// Config contains the configuration of acmeproxy's public server.
type Config struct {
	ACMEDirectoryURL string
	HTTPAPIAddr      string
	DataDir          string
	Logger           log.Logger
}

// Server is acmeproxy's public server.
type Server struct {
	httpAPIServer *httpServer
	acmeAgent     *acme.Agent
	boltDB        *db.Bolt
	logger        log.Logger
}

// New creates a new Server and configures it using cfg.
func New(cfg Config) (*Server, error) {
	boltDB := &db.Bolt{
		FilePath: filepath.Join(cfg.DataDir, "acmeproxy.db"),
		FileMode: 0600,
	}
	acmeClient := &acmeclient.Client{
		DirectoryURL: cfg.ACMEDirectoryURL,
		HTTP01Solver: acmeclient.NewHTTP01Solver(),
	}
	acmeAgent := &acme.Agent{
		Domains:      boltDB.DomainRepository(),
		Clients:      boltDB.ClientRepository(),
		Certificates: acmeClient,
		ACMEAccounts: acmeClient,
	}
	httpAPIServer := &httpServer{
		Addr:   cfg.HTTPAPIAddr,
		Logger: cfg.Logger,
		Handler: httpapi.NewRouter(httpapi.Config{
			Solver: acmeClient.HTTP01Solver,
		}),
	}
	return &Server{
		httpAPIServer: httpAPIServer,
		acmeAgent:     acmeAgent,
		boltDB:        boltDB,
		logger:        cfg.Logger,
	}, nil
}

// Start starts the Server without blocking.
func (s *Server) Start() {
	err := s.boltDB.Open()
	if err != nil {
		// TODO (fhofherr) panic is not an option for v0.1.0. But it is good enough for testing
		panic("open bolt db")
	}
	s.httpAPIServer.Start()
	s.registerAcmeproxyDomain()
}

// Shutdown performs a graceful shutdown of the Server.
func (s *Server) Shutdown(ctx context.Context) {
	s.httpAPIServer.Shutdown(ctx)
	s.boltDB.Close()
}

func (s *Server) registerAcmeproxyDomain() {
	tmpClientID := uuid.Must(uuid.NewRandom())
	// TODO (fhofherr) make acmeproxy admin mail configurable
	err := s.acmeAgent.RegisterClient(tmpClientID, "")
	if err != nil {
		log.Log(s.logger,
			"level", "error",
			"error", err,
			"message", "register default client")
	}
	// TODO (fhofherr) make acmeproxy domain configurable
	err = s.acmeAgent.RegisterDomain(tmpClientID, "www.example.com")
	if err != nil {
		log.Log(s.logger,
			"level", "error",
			"error", err,
			"message", fmt.Sprintf("register acmeproxy domain: %s", "www.example.com"))
	}
}
