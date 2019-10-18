package httpapi

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

// HandlerFactory creates handlers for use by the HTTP API's router.
type HandlerFactory interface {
	Handler(func(*http.Request) map[string]string) http.Handler
}

// Config configures the public, non-encrypted HTTP API of acmeproxy.
type Config struct {
	Solver HandlerFactory // Presents solutions to HTTP01 challenges to ACME CA.
}

// NewRouter creates an http.Handler which serves acmeproxy's public,
// non-encrypted HTTP API.
func NewRouter(cfg Config) http.Handler {
	r := chi.NewRouter()

	r.Get(
		"/.well-known/acme-challenge/{token}",
		cfg.Solver.Handler(acmeChallengeParams).ServeHTTP,
	)

	return r
}

func acmeChallengeParams(req *http.Request) map[string]string {
	return map[string]string{
		"domain": strings.Split(req.Host, ":")[0],
		"token":  chi.URLParam(req, "token"),
	}
}
