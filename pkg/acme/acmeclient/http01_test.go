package acmeclient_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/stretchr/testify/assert"
)

func TestHTTP01Solver_GetKeyAuthForTokenAndDomain(t *testing.T) {
	domain := "www.example.com"
	token := "token"
	keyAuth := "keyAuth"

	solver := acmeclient.HTTP01Solver{}
	err := solver.Present(domain, token, keyAuth)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler := solver.Handler(newParamExtractor(t, req, domain, token))
	status, actualKeyAuth := exerciseHandler(t, handler)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, keyAuth, actualKeyAuth)
}

func TestHTTP01Solver_ReturnNotFoundOnMissingKeyAuth(t *testing.T) {
	domain := "www.example.com"
	token := "token"

	solver := acmeclient.HTTP01Solver{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler := solver.Handler(newParamExtractor(t, req, domain, token))
	status, _ := exerciseHandler(t, handler)

	assert.Equal(t, http.StatusNotFound, status)
}

func TestHTTP01Solver_CleanUpRemovesKeyAuth(t *testing.T) {
	domain := "www.example.com"
	token := "token"
	keyAuth := "keyAuth"

	solver := acmeclient.HTTP01Solver{}
	err := solver.Present(domain, token, keyAuth)
	assert.NoError(t, err)

	err = solver.CleanUp(domain, token, keyAuth)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler := solver.Handler(newParamExtractor(t, req, domain, token))
	status, _ := exerciseHandler(t, handler)

	assert.Equal(t, http.StatusNotFound, status)
}

func TestHTTP01Solver_CleanUpOnNewSolverDoesNotFail(t *testing.T) {
	domain := "www.example.com"
	token := "token"
	keyAuth := "keyAuth"

	solver := acmeclient.HTTP01Solver{}
	err := solver.CleanUp(domain, token, keyAuth)
	assert.NoError(t, err)
}

func TestHTTP01Solver_ConcurrentAccess(t *testing.T) {
	solver := acmeclient.HTTP01Solver{}
	n := 10
	maxSleep := int64(31)
	wg := sync.WaitGroup{}
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			domain := fmt.Sprintf("www.example%d.com", i)
			token := fmt.Sprintf("token%d", i)
			keyAuth := fmt.Sprintf("keyAuth%d", i)

			time.Sleep(time.Duration(rand.Int63n(maxSleep)) * time.Millisecond)

			err := solver.Present(domain, token, keyAuth)
			assert.NoError(t, err)

			time.Sleep(time.Duration(rand.Int63n(maxSleep)) * time.Millisecond)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			handler := solver.Handler(newParamExtractor(t, req, domain, token))
			status, keyAuth := exerciseHandler(t, handler)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, status)
			assert.Equal(t, keyAuth, keyAuth)

			time.Sleep(time.Duration(rand.Int63n(maxSleep)) * time.Millisecond)

			err = solver.CleanUp(domain, token, keyAuth)
			assert.NoError(t, err)

			wg.Done()
		}(i)
	}
	wg.Wait()
}

func newParamExtractor(t *testing.T, req *http.Request, domain, token string) func(*http.Request) map[string]string {
	return func(actual *http.Request) map[string]string {
		if !reflect.DeepEqual(req, actual) {
			t.Fatal("Actual request did not equal expected")
			return nil
		}
		return map[string]string{
			"domain": domain,
			"token":  token,
		}
	}
}

func exerciseHandler(t *testing.T, handler http.Handler) (int, string) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return rr.Code, string(body)
}
