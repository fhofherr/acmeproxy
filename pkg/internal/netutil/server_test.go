package netutil_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/internal/netutil"
	"github.com/stretchr/testify/assert"
)

func TestListenAndServe_StartServer(t *testing.T) {
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	}
	defer s.Close()
	addrC := make(chan string)
	go netutil.ListenAndServe(s, netutil.NotifyAddr(addrC)) // nolint: errcheck
	addr := netutil.GetAddr(t, addrC)
	res, err := http.Get(fmt.Sprintf("http://%s/", addr))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestListenAndServe_ReturnsListenerErrors(t *testing.T) {
	s := &http.Server{}
	defer s.Close()

	addrC := make(chan string)
	go netutil.ListenAndServe(s, netutil.NotifyAddr(addrC)) // nolint: errcheck

	addr := netutil.GetAddr(t, addrC)
	s2 := &http.Server{}
	err := netutil.ListenAndServe(s2, netutil.WithAddr(addr))
	assert.Error(t, err)
}

func TestListenAndServe_IgnoresErrServerClosed(t *testing.T) {
	s := &http.Server{}
	addrC := make(chan string)
	errC := make(chan error)
	go func(errC chan<- error) {
		errC <- netutil.ListenAndServe(s, netutil.NotifyAddr(addrC))
	}(errC)

	// We are just interested in the address being sent. This signals that
	// the server is listening and ready to accept connections.
	_ = netutil.GetAddr(t, addrC)

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	err := netutil.GetErr(t, errC)
	assert.NoError(t, err)
}

func TestListenAndServe_ReturnsServerErrors(t *testing.T) {
	s := &http.Server{}
	defer s.Close()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addrC := make(chan string)
	errC := make(chan error)
	go func(errC chan<- error) {
		errC <- netutil.ListenAndServe(s, netutil.NotifyAddr(addrC), netutil.WithListener(l))
	}(errC)

	addr := netutil.GetAddr(t, addrC)
	l.Close()

	if res, err := http.Get(fmt.Sprintf("http://%s/", addr)); err == nil {
		// we actually don't expect this to work, since we closed l before
		// making the request.
		defer res.Body.Close()
		t.Fatal("Received a resonse from s")
	}
	assert.Error(t, netutil.GetErr(t, errC))
}
