package httpapi_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/httpapi"
	"github.com/stretchr/testify/assert"
)

func TestHTTP01ChallengeEndpoint(t *testing.T) {
	token := "some-token"
	domain := "www.example.com"

	solverFactory := &httpapi.MockHandlerFactory{}
	cfg := httpapi.Config{Solver: solverFactory}
	router := httpapi.NewRouter(cfg)

	url := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s", domain, token)
	assert.HTTPSuccess(t, router.ServeHTTP, http.MethodGet, url, nil)
	assert.Len(t, solverFactory.Params, 1)
	assert.Equal(t, domain, solverFactory.Params[0]["domain"])
	assert.Equal(t, token, solverFactory.Params[0]["token"])
}
