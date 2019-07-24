package httpapi

import (
	"net/http"

	"github.com/go-chi/chi"
)

type Server struct {
	ChallengeSolver ChallengeSolver
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
