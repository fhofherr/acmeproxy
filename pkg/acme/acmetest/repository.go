package acmetest

import (
	"sync"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// InMemoryDomainRepository is a simple in-memory implementation of
// acme.DomainRepository
type InMemoryDomainRepository struct {
	domains map[string]acme.Domain
	mu      sync.Mutex
}

// UpdateDomain saves or updates a domain in the InMemoryDomainRepository using
// the update function f.
//
// UpdateDomain passes a pointer to an domain object to f. If f does not return
// an error, UpdateDomain stores the updated domain object.
//
// Callers must not hold on to the *acme.Domain parameter passed to f. Rather
// they should use the domain returned by UpdateDomain. If UpdateDomain returns
// an error, the result of f is not stored in the repository.
func (r *InMemoryDomainRepository) UpdateDomain(domainName string, f func(*acme.Domain) error) (acme.Domain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.domains == nil {
		r.domains = make(map[string]acme.Domain)
	}
	domain := r.domains[domainName]
	err := f(&domain)
	if err != nil {
		return acme.Domain{}, errors.Wrapf(err, "update domain: %s", domainName)
	}
	r.domains[domainName] = domain
	return domain, nil
}

// GetDomain retrieves a domain from the repository
func (r *InMemoryDomainRepository) GetDomain(domainName string) (acme.Domain, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.domains == nil {
		return acme.Domain{}, acme.DomainNotFound{DomainName: domainName}
	}
	if domain, ok := r.domains[domainName]; ok {
		return domain, nil
	}
	return acme.Domain{}, acme.DomainNotFound{DomainName: domainName}
}

// InMemoryClientRepository is a simple in-memory implementation of
// acme.ClientRepository.
type InMemoryClientRepository struct {
	clients map[uuid.UUID]acme.Client
	mu      sync.Mutex
}

// UpdateClient saves or updates a client in the InMemoryClientRepository using
// the update function f.
//
// UpdateClient passes a pointer to a client object to f. If f does not return
// an error, UpdateClient stores the updated client.
//
// Callers must not hold on to the *acme.Client parameter passed to f. Rather
// they should use the client returned by UpdateClient. If UpdateClient returns
// an error, the result of f is not stored in the repository.
func (r *InMemoryClientRepository) UpdateClient(id uuid.UUID, f func(*acme.Client) error) (acme.Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.clients == nil {
		r.clients = make(map[uuid.UUID]acme.Client)
	}
	client := r.clients[id]
	err := f(&client)
	if err != nil {
		return acme.Client{}, errors.Wrapf(err, "update client: %v", id)
	}
	r.clients[id] = client
	return client, nil
}

// GetClient retrieves the client from the client repository
func (r *InMemoryClientRepository) GetClient(id uuid.UUID) (acme.Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.clients == nil {
		return acme.Client{}, acme.ClientNotFound{ClientID: id}
	}
	if c, ok := r.clients[id]; ok {
		return c, nil
	}
	return acme.Client{}, acme.ClientNotFound{ClientID: id}

}
