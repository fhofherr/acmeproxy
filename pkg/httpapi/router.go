package httpapi

import (
	"net/http"

	"github.com/go-chi/chi"
)

// Config configures the public, non-encrypted HTTP API of acmeproxy.
type Config struct {
	Solver ChallengeSolver // Presents solutions to HTTP01 challenges to ACME CA.
}

// NewRouter creates an http.Handler which serves acmeproxy's public,
// non-encrypted HTTP API.
func NewRouter(cfg Config) http.Handler {
	r := chi.NewRouter()

	r.Get("/.well-known/acme-challenge/{token}",
		challengeHandler{
			Params: urlParams("token"),
			Solver: cfg.Solver,
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
