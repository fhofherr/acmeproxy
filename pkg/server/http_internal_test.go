package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"

	"github.com/stretchr/testify/assert"
)

func TestStartNonEncryptedHTTPServer(t *testing.T) {
	// Choose a random port number between 49152 and 65535. We can't let the
	// operating system choose the port vor us (by using 0 as the port numer)
	// as we need the port to connect to our server afterwards. The http.Server
	// does not provide a way to obtain the server's address if the port number
	// was chosen by the os.
	port := rand.Intn(16383) + 49152
	message := "Server was called"
	server := &httpServer{
		Addr: fmt.Sprintf("127.0.0.1:%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte(message))
		}),
	}
	go func() {
		assert.NoError(t, server.Start())
	}()
	defer func() {
		assert.NoError(t, server.Shutdown(context.Background()))
	}()
	url := fmt.Sprintf("http://%s/", server.Addr)

	var (
		statusCode int
		body       []byte
	)
	testsupport.Retry(t, 3, 5*time.Millisecond, func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		statusCode = resp.StatusCode
		body, err = ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		// We don't want to retry if reading the body failed.
		return nil
	})
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, message, string(body))
}

func TestShutdownOfUnstartedServer(t *testing.T) {
	tmpl := errors.New(errors.Op("server/httpServer.Shutdown"), "not started")
	server := &httpServer{}
	err := server.Shutdown(context.Background())
	assert.True(t, errors.Match(tmpl, err))
}
