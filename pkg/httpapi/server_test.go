package httpapi_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/httpapi"
	"github.com/fhofherr/acmeproxy/pkg/httpapi/httpapitest"
	"github.com/stretchr/testify/assert"
)

func TestServeHTTP01Challenge(t *testing.T) {
	token := "some-token"
	keyAuth := "key-auth"
	domain := "127.0.0.1"
	port := "8080"

	solver := &httpapitest.MockChallengeSolver{}
	solver.Test(t)
	solver.
		On("SolveChallenge", domain, token).
		Return(keyAuth, nil)

	server := httpapi.Server{
		Addr:            fmt.Sprintf("%s:%s", domain, port),
		ChallengeSolver: solver,
	}
	server.Start()
	defer server.Shutdown(context.Background())

	url := fmt.Sprintf("http://%s:%s/.well-known/acme-challenge/%s",
		domain, port, token)
	resp, err := http.Get(url)

	if assert.NoError(t, err) {
		defer resp.Body.Close()

		assert.Equal(t, resp.StatusCode, http.StatusOK)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, keyAuth, string(body))
	}

	solver.AssertCalled(t, "SolveChallenge", domain, token)
}
