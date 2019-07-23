package acmetest

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
)

// NewChallengeServer creates an httptest.Server which uses the handler to
// serve HTTP01 challenges.
func NewChallengeServer(t *testing.T, handler *acme.HTTP01Handler, port int) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		portSuffix := fmt.Sprintf(":%d", port)
		domain := strings.Replace(req.Host, portSuffix, "", -1)
		pathParts := strings.Split(req.URL.Path, "/")
		if len(pathParts) == 0 {
			t.Fatalf("Could not obtain token from url: %v", req.URL)
		}
		token := pathParts[len(pathParts)-1]
		keyAuth, err := handler.HandleChallenge(domain, token)
		w.Header().Add("content-type", "application/octet-stream")
		if failedErr, ok := err.(acme.ErrChallengeFailed); ok {
			t.Log(failedErr)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			t.Errorf("HandleChallenge failed: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte(keyAuth))
		if err != nil {
			t.Error(err)
		}
	}))

	// To enable the dev environment started in docker-compose to access to
	// our test server started on the docker-compose host machine, we have to
	// make it listen on all ports.
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
