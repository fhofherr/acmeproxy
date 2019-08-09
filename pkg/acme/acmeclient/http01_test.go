package acmeclient_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/stretchr/testify/assert"
)

func TestHTTP01Handler_GetKeyAuthForTokenAndDomain(t *testing.T) {
	domain := "www.example.com"
	token := "token"
	keyAuth := "keyAuth"

	handler := acmeclient.NewHTTP01Solver()
	err := handler.Present(domain, token, keyAuth)
	assert.NoError(t, err)
	actualKeyAuth, err := handler.SolveChallenge(domain, token)
	assert.NoError(t, err)
	assert.Equal(t, keyAuth, actualKeyAuth)
}

func TestHTTP01Handler_ReturnErrorOnMissingKeyAuth(t *testing.T) {
	domain := "www.example.com"
	token := "token"

	handler := acmeclient.NewHTTP01Solver()
	_, err := handler.SolveChallenge(domain, token)
	assert.Error(t, err)
	assert.Equal(t, acmeclient.ErrChallengeFailed{Domain: domain, Token: token}, err)
}

func TestHTTP01Handler_CleanUpAfterSuccessfulChallenge(t *testing.T) {
	domain := "www.example.com"
	token := "token"
	keyAuth := "keyAuth"

	handler := acmeclient.NewHTTP01Solver()
	err := handler.Present(domain, token, keyAuth)
	assert.NoError(t, err)

	err = handler.CleanUp(domain, token, keyAuth)
	assert.NoError(t, err)

	_, err = handler.SolveChallenge(domain, token)
	assert.Equal(t, acmeclient.ErrChallengeFailed{Domain: domain, Token: token}, err)
}

func TestHTTP01Handler_ConcurrentAccess(t *testing.T) {
	handler := acmeclient.NewHTTP01Solver()
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

			err := handler.Present(domain, token, keyAuth)
			assert.NoError(t, err)

			time.Sleep(time.Duration(rand.Int63n(maxSleep)) * time.Millisecond)

			act, err := handler.SolveChallenge(domain, token)
			assert.NoError(t, err)
			assert.Equal(t, keyAuth, act)

			time.Sleep(time.Duration(rand.Int63n(maxSleep)) * time.Millisecond)

			err = handler.CleanUp(domain, token, keyAuth)
			assert.NoError(t, err)

			wg.Done()
		}(i)
	}
	wg.Wait()
}