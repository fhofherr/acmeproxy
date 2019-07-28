package server

import (
	"context"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/httpapi"
	"github.com/fhofherr/golf/log"
)

// Config contains the configuration of acmeproxy's public server.
type Config struct {
	ACMEAgentCfg acme.AgentConfig
	HTTPAPIAddr  string

	Logger log.Logger
}

// Server is acmeproxy's public server.
type Server struct {
	httpAPIServer *httpServer
}

// New creates a new Server and configures it using cfg.
func New(cfg Config) *Server {
	acmeAgent := acme.NewAgent(cfg.ACMEAgentCfg)
	httpAPIServer := &httpServer{
		Addr:   cfg.HTTPAPIAddr,
		Logger: cfg.Logger,
		Handler: httpapi.NewRouter(httpapi.Config{
			Solver: acmeAgent.HTTP01ChallengeSolver(),
		}),
	}
	return &Server{
		httpAPIServer: httpAPIServer,
	}
}

// Start starts the Server without blocking.
func (s *Server) Start() {
	s.httpAPIServer.Start()
}

// Shutdown performs a graceful shutdown of the Server.
func (s *Server) Shutdown(ctx context.Context) {
	s.httpAPIServer.Shutdown(ctx)
}
