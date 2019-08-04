package acmetest

import (
	"crypto"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/stretchr/testify/assert"
)

type InMemoryAccountCreator struct {
	accounts map[string]*fakeAccountData
	mu       sync.Mutex
}

// CreateAccount creates an random account URL and returns it.
//
// Additionally the InMemoryAccountCreator remembers the account URL and the
// private key that was passed when creating it.
func (ac *InMemoryAccountCreator) CreateAccount(key crypto.PrivateKey, email string) (string, error) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.accounts == nil {
		ac.accounts = make(map[string]*fakeAccountData)
	}
	accountURL := fmt.Sprintf("https://acme.example.com/directory/%d", rand.Int())
	ac.accounts[accountURL] = &fakeAccountData{
		Key:   key,
		Email: email,
	}
	return accountURL, nil
}

// AssertCreated asserts that this InMemoryAccountCreator created an account for
// client.
func (ac *InMemoryAccountCreator) AssertCreated(t *testing.T, email string, client acme.Client) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.accounts == nil {
		t.Error("InMemoryAccountCreator has no accounts")
		return
	}
	data, ok := ac.accounts[client.AccountURL]
	if !ok {
		t.Errorf("InMemoryAccountCreator did not create an account for %s", client.AccountURL)
		return
	}
	assert.Equalf(t, data.Key, client.Key, "Key of client %s did not match stored key", client.AccountURL)
	assert.Equalf(t, data.Email, email, "Email of client %s did not match stored email", client.AccountURL)

}

type fakeAccountData struct {
	Key   crypto.PrivateKey
	Email string
}
