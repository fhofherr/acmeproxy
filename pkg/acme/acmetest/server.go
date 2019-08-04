package acmetest

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
)

// NewChallengeServer creates an httptest.Server which uses the handler to
// serve HTTP01 challenges.
func NewChallengeServer(t *testing.T, handler *acmeclient.HTTP01Solver, port int) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		portSuffix := fmt.Sprintf(":%d", port)
		domain := strings.Replace(req.Host, portSuffix, "", -1)
		pathParts := strings.Split(req.URL.Path, "/")
		if len(pathParts) == 0 {
			t.Fatalf("Could not obtain token from url: %v", req.URL)
		}
		token := pathParts[len(pathParts)-1]
		keyAuth, err := handler.SolveChallenge(domain, token)
		w.Header().Add("content-type", "application/octet-stream")
		if failedErr, ok := err.(acmeclient.ErrChallengeFailed); ok {
			t.Log(failedErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("SolveChallenge failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(keyAuth))
		if err != nil {
			t.Error(err)
		}
	}))

	// The dev environment is started in docker-compose or in separate
	// Docker containers by our CI server. In order for it to be able to access
	// to the test server started on the host machine or a separate Docker
	// container we have to make it listen on all interfaces.
	//
	// If the host machine has a firewall it has to temporarily allow access
	// to the dev server through the firewall.
	address := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	server.Listener = listener
	server.Start()
	return server
}
