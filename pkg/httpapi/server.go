package httpapi

import (
	"context"
	"net/http"

	"github.com/fhofherr/golf/log"
	"github.com/go-chi/chi"
)

// Server serves the HTTP API's non-encrypted endpoints.
type Server struct {
	Addr            string
	ChallengeSolver ChallengeSolver
	Logger          log.Logger
	httpServer      *http.Server
}

// Start starts the server.
func (s *Server) Start() {
	handler := newRouter(s)
	s.httpServer = &http.Server{
		Addr:    s.Addr,
		Handler: handler,
	}
	go s.listenAndServe()
	// TODO consider polling an ready enpoint to signal once the server is ready
}

// Shutdown performs a graceful shutdown of the server.
func (s *Server) Shutdown(ctx context.Context) {
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Log(s.Logger, "level", "error", "error", err)
	}
}

func (s *Server) listenAndServe() {
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Log(s.Logger, "level", "error", "error", err)
	}
}

func newRouter(server *Server) http.Handler {
	r := chi.NewRouter()

	r.Get("/.well-known/acme-challenge/{token}",
		challengeHandler{
			Params: urlParams("token"),
			Solver: server.ChallengeSolver,
		}.ServeHTTP)

	return r
}

type paramExtractor func(req *http.Request) map[string]string

func urlParams(ks ...string) paramExtractor {
	return func(req *http.Request) map[string]string {
		params := make(map[string]string, len(ks))
		for _, k := range ks {
			params[k] = chi.URLParam(req, k)
		}
		return params
	}
}
