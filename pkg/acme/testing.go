package acme

import (
	"crypto"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// InMemoryAccountCreator is used to create fake letsencrypt accounts.
// It remembers the account URL and the private key that was passed when
// creating a fake account.
type InMemoryAccountCreator struct {
	accounts map[string]*fakeAccountData
	mu       sync.Mutex
}

// CreateAccount creates an random account URL and returns it.
func (ac *InMemoryAccountCreator) CreateAccount(key crypto.PrivateKey, email string) (string, error) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.accounts == nil {
		ac.accounts = make(map[string]*fakeAccountData)
	}
	accountURL := fmt.Sprintf("https://acme.example.com/directory/%d", rand.Int()) //nolint:gosec
	ac.accounts[accountURL] = &fakeAccountData{
		Key:   key,
		Email: email,
	}
	return accountURL, nil
}

// AssertCreated asserts that this InMemoryAccountCreator created an account for
// client.
func (ac *InMemoryAccountCreator) AssertCreated(t *testing.T, email string, client Client) {
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

// InMemoryDomainRepository is a simple in-memory implementation of
// DomainRepository
type InMemoryDomainRepository struct {
	domains map[string]Domain
	mu      sync.Mutex
}

// UpdateDomain saves or updates a domain in the InMemoryDomainRepository using
// the update function f.
//
// UpdateDomain passes a pointer to an domain object to f. If f does not return
// an error, UpdateDomain stores the updated domain object.
//
// Callers must not hold on to the *Domain parameter passed to f. Rather
// they should use the domain returned by UpdateDomain. If UpdateDomain returns
// an error, the result of f is not stored in the repository.
func (r *InMemoryDomainRepository) UpdateDomain(domainName string, f func(*Domain) error) (Domain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.domains == nil {
		r.domains = make(map[string]Domain)
	}
	domain := r.domains[domainName]
	err := f(&domain)
	if err != nil {
		return Domain{}, errors.Wrapf(err, "update domain: %s", domainName)
	}
	r.domains[domainName] = domain
	return domain, nil
}

// GetDomain retrieves a domain from the repository
func (r *InMemoryDomainRepository) GetDomain(domainName string) (Domain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.domains == nil {
		return Domain{}, DomainNotFound{DomainName: domainName}
	}
	if domain, ok := r.domains[domainName]; ok {
		return domain, nil
	}
	return Domain{}, DomainNotFound{DomainName: domainName}
}

// InMemoryClientRepository is a simple in-memory implementation of
// ClientRepository.
type InMemoryClientRepository struct {
	clients map[uuid.UUID]Client
	mu      sync.Mutex
}

// UpdateClient saves or updates a client in the InMemoryClientRepository using
// the update function f.
//
// UpdateClient passes a pointer to a client object to f. If f does not return
// an error, UpdateClient stores the updated client.
//
// Callers must not hold on to the *Client parameter passed to f. Rather
// they should use the client returned by UpdateClient. If UpdateClient returns
// an error, the result of f is not stored in the repository.
func (r *InMemoryClientRepository) UpdateClient(id uuid.UUID, f func(*Client) error) (Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.clients == nil {
		r.clients = make(map[uuid.UUID]Client)
	}
	client := r.clients[id]
	err := f(&client)
	if err != nil {
		return Client{}, errors.Wrapf(err, "update client: %v", id)
	}
	r.clients[id] = client
	return client, nil
}

// GetClient retrieves the client from the client repository
func (r *InMemoryClientRepository) GetClient(id uuid.UUID) (Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.clients == nil {
		return Client{}, ClientNotFound{ClientID: id}
	}
	if c, ok := r.clients[id]; ok {
		return c, nil
	}
	return Client{}, ClientNotFound{ClientID: id}
}
