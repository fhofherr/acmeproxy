package httpapi_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/api/httpapi"
	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"

	"github.com/stretchr/testify/assert"
)

func TestServer_ServeHTTP01ChallengeEndpoint(t *testing.T) {
	token := "some-token"
	domain := "127.0.0.1"

	solverFactory := &httpapi.MockHandlerFactory{}
	server := &httpapi.Server{
		Solver: solverFactory,
	}
	addrC := make(chan string)
	go netutil.ListenAndServe(server, netutil.NotifyAddr(addrC), netutil.WithAddr(domain+":0")) // nolint: errcheck
	addr := netutil.GetAddr(t, addrC)
	defer server.Shutdown(context.Background()) //nolint: errcheck

	url := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s", addr, token)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Len(t, solverFactory.Params, 1)
	assert.Equal(t, domain, solverFactory.Params[0]["domain"])
	assert.Equal(t, token, solverFactory.Params[0]["token"])
}

func TestServer_ServeHealthEndpoint(t *testing.T) {
	server := &httpapi.Server{
		Solver: &httpapi.MockHandlerFactory{},
	}
	addrC := make(chan string)
	go netutil.ListenAndServe(server, netutil.NotifyAddr(addrC), netutil.WithAddr("127.0.0.1:0")) // nolint: errcheck
	addr := netutil.GetAddr(t, addrC)
	defer server.Shutdown(context.Background()) // nolint: errcheck

	resp, err := http.Get(fmt.Sprintf("http://%s/health", addr))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, resp.Header.Get("content-type"), "application/health+json")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, `{"status": "pass"}`, string(body))
}

func TestServer_CannotServeWithoutSolver(t *testing.T) {
	errC := make(chan error)
	server := &httpapi.Server{}
	go func(errC chan<- error) {
		errC <- netutil.ListenAndServe(server)
	}(errC)
	assert.Error(t, netutil.GetErr(t, errC))
}

func TestServer_CannotShutdownUnstartedServer(t *testing.T) {
	server := &httpapi.Server{
		Solver: &httpapi.MockHandlerFactory{},
	}
	err := server.Shutdown(context.Background())
	assert.Error(t, err)
}
