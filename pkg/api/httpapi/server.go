package httpapi

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/golf/log"
	"github.com/go-chi/chi"
)

// HandlerFactory creates handlers for use by the HTTP API's router.
type HandlerFactory interface {
	Handler(func(*http.Request) map[string]string) http.Handler
}

// Server serves the public, non-encrypted part of acmeproxy's http API.
type Server struct {
	Solver     HandlerFactory // Presents solutions to HTTP01 challenges to ACME CA.
	Logger     log.Logger
	httpServer *http.Server
	once       sync.Once
}

func (s *Server) initialize() error {
	const op errors.Op = "httpapi/server.initialize"
	var err error

	s.once.Do(func() {
		if s.Solver == nil {
			err = errors.New(op, "no solver set")
			return
		}
		s.httpServer = &http.Server{
			Handler: s.newRouter(),
		}
	})

	return err
}

// Serve accepts incomming connections on the listener l.
//
// Server initialized the Server if this has not been done yet. If the
// initialization fails Serve returns an error. Likewise Serve returns any error
// that occurs while accpeting incomming connections.
//
// Since Server uses net/http.Server internally and Serve delegates to
// net/http.Server.Serve it will return net/http.ErrServerClosed once the
// Server is shut down.
func (s *Server) Serve(l net.Listener) error {
	const op errors.Op = "httpapi/server.Serve"

	if err := s.initialize(); err != nil {
		return errors.New(op, err)
	}
	return errors.Wrap(s.httpServer.Serve(l), op, "serve http")
}

// Shutdown gracefully stops the Server.
func (s *Server) Shutdown(ctx context.Context) error {
	const op errors.Op = "httpapi/server.Shutdown"

	if s.httpServer == nil {
		return errors.New(op, "not started")
	}
	return errors.Wrap(s.httpServer.Shutdown(ctx), op, "shutdown")
}

func (s *Server) newRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", s.healthHandler)
	r.Get(
		"/.well-known/acme-challenge/{token}",
		s.Solver.Handler(acmeChallengeParams).ServeHTTP,
	)

	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "application/health+json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status": "pass"}`))
	errors.Log(s.Logger, err)
}

func acmeChallengeParams(req *http.Request) map[string]string {
	return map[string]string{
		"domain": strings.Split(req.Host, ":")[0],
		"token":  chi.URLParam(req, "token"),
	}
}
