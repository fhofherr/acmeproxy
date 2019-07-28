package server

import (
	"context"
	"net/http"

	"github.com/fhofherr/golf/log"
)

// httpServer serves acmeproxy's public, non-encrypted httpServer endpoints.
type httpServer struct {
	Addr       string
	Logger     log.Logger
	Handler    http.Handler
	httpServer *http.Server
}

// Start starts the server.
func (s *httpServer) Start() {
	s.httpServer = &http.Server{
		Addr:    s.Addr,
		Handler: s.Handler,
	}
	go s.listenAndServe()
}

// Shutdown performs a graceful shutdown of the server.
func (s *httpServer) Shutdown(ctx context.Context) {
	if s.httpServer == nil {
		log.Log(s.Logger,
			"level", "warn",
			"message", "Server not started. Skipping shutdown sequence")
		return
	}
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Log(s.Logger, "level", "error", "error", err)
	}
}

func (s *httpServer) listenAndServe() {
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Log(s.Logger, "level", "error", "error", err)
	}
}
