package httpapi_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/httpapi"
	"github.com/fhofherr/acmeproxy/pkg/httpapi/httpapitest"
	"github.com/stretchr/testify/assert"
)

func TestHTTP01ChallengeEndpoint(t *testing.T) {
	token := "some-token"
	keyAuth := "key-auth"
	domain := "www.example.com"

	solver := &httpapitest.MockChallengeSolver{}
	solver.Test(t)
	solver.
		On("SolveChallenge", domain, token).
		Return(keyAuth, nil)
	cfg := httpapi.Config{Solver: solver}
	router := httpapi.NewRouter(cfg)

	url := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s", domain, token)
	assert.HTTPSuccess(t, router.ServeHTTP, http.MethodGet, url, nil)
	solver.AssertCalled(t, "SolveChallenge", domain, token)
}
