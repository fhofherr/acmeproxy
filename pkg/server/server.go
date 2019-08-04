package server

import (
	"context"
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmetest"
	"github.com/fhofherr/acmeproxy/pkg/httpapi"
	"github.com/fhofherr/golf/log"
	"github.com/google/uuid"
)

// Config contains the configuration of acmeproxy's public server.
type Config struct {
	ACMEDirectoryURL string
	HTTPAPIAddr      string
	Logger           log.Logger
}

// Server is acmeproxy's public server.
type Server struct {
	httpAPIServer *httpServer
	acmeAgent     *acme.Agent
	logger        log.Logger
}

// New creates a new Server and configures it using cfg.
func New(cfg Config) (*Server, error) {
	acmeClient := &acmeclient.Client{
		DirectoryURL: cfg.ACMEDirectoryURL,
		HTTP01Solver: acmeclient.NewHTTP01Solver(),
	}
	acmeAgent := &acme.Agent{
		Domains:      &acmetest.InMemoryDomainRepository{},
		Clients:      &acmetest.InMemoryClientRepository{},
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
		logger:        cfg.Logger,
	}, nil
}

// Start starts the Server without blocking.
func (s *Server) Start() {
	s.httpAPIServer.Start()
	s.registerAcmeproxyDomain()
}

// Shutdown performs a graceful shutdown of the Server.
func (s *Server) Shutdown(ctx context.Context) {
	s.httpAPIServer.Shutdown(ctx)
}

func (s *Server) registerAcmeproxyDomain() {
	tmpClientID := uuid.Must(uuid.NewRandom())
	// TODO(fhofherr): make acmeproxy admin mail configurable
	err := s.acmeAgent.RegisterClient(tmpClientID, "")
	if err != nil {
		log.Log(s.logger,
			"level", "error",
			"error", err,
			"message", "register default client")
	}
	// TODO(fhofherr): make acmeproxy domain configurable
	err = s.acmeAgent.RegisterDomain(tmpClientID, "www.example.com")
	if err != nil {
		log.Log(s.logger,
			"level", "error",
			"error", err,
			"message", fmt.Sprintf("register acmeproxy domain: %s", "www.example.com"))
	}
}
