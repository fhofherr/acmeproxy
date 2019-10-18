package server

import (
	"context"
	"net/http"

	"github.com/fhofherr/acmeproxy/pkg/errors"
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
//
// Calling start blocks the calling go routine.
func (s *httpServer) Start() error {
	const op errors.Op = "server/httpServer.Start"

	s.httpServer = &http.Server{
		Addr:    s.Addr,
		Handler: s.Handler,
	}
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return errors.New(op, err)
	}
	return nil
}

// Shutdown performs a graceful shutdown of the server.
func (s *httpServer) Shutdown(ctx context.Context) error {
	const op errors.Op = "server/httpServer.Shutdown"

	if s.httpServer == nil {
		return errors.New(op, "not started")
	}
	return errors.Wrap(s.httpServer.Shutdown(ctx), op)
}
