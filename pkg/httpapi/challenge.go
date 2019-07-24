package httpapi

import (
	"net/http"
	"strings"

	"github.com/fhofherr/golf/log"
)

// ChallengeSolver wraps the SolveChallenge method which tries to solve an
// HTTP01 challenge for the provided domain and token.
//
// If the SolveChallenge method was not able to solve the challenge for domain
// and token it should return an error implementing the ErrChallengeFailed
// interface.
//
// SolveChallenge may return other errors for any additional reasons.
type ChallengeSolver interface {
	SolveChallenge(domain, token string) (string, error)
}

// ErrChallengeFailed wraps the ChallengeFailed method.
//
// This interface should be implemented by errors wishing to signal that solving
// the challenge failed for non-technical reasons.
//
// ChallengeFailed may return false to signal that the implementor does not wish
// this error to be treated in a special way.
type ErrChallengeFailed interface {
	ChallengeFailed() bool
}

func asErrChallengeFailed(err error) (ErrChallengeFailed, bool) {
	if cfErr, ok := err.(ErrChallengeFailed); ok {
		return cfErr, cfErr.ChallengeFailed()
	}
	return nil, false
}

type challengeHandler struct {
	Params paramExtractor
	Solver ChallengeSolver
	Logger log.Logger
}

func (h challengeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/octet-stream")

	ps := h.Params(r)
	token := ps["token"]
	domain := strings.Split(r.Host, ":")[0]
	keyAuth, err := h.Solver.SolveChallenge(domain, token)
	if _, ok := asErrChallengeFailed(err); ok {
		log.Log(h.Logger, "level", "info", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Log(h.Logger, "level", "error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(keyAuth))
	if err != nil {
		log.Log(h.Logger, "level", "warn", "error", err)
	}
}
