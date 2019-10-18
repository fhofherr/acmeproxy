package acme

import (
	"crypto"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"sync"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
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
// user.
func (ac *InMemoryAccountCreator) AssertCreated(t *testing.T, email string, user User) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.accounts == nil {
		t.Error("InMemoryAccountCreator has no accounts")
		return
	}
	data, ok := ac.accounts[user.AccountURL]
	if !ok {
		t.Errorf("InMemoryAccountCreator did not create an account for %s", user.AccountURL)
		return
	}
	assert.Equalf(t, data.Key, user.Key, "Key of user %s did not match stored key", user.AccountURL)
	assert.Equalf(t, data.Email, email, "Email of user %s did not match stored email", user.AccountURL)
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
	const op errors.Op = "acme/inMemoryDomainRepository.UpdateDomain"

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.domains == nil {
		r.domains = make(map[string]Domain)
	}
	domain := r.domains[domainName]
	err := f(&domain)
	if err != nil {
		return Domain{}, errors.New(op, fmt.Sprintf("update domain: %s", domainName), err)
	}
	r.domains[domainName] = domain
	return domain, nil
}

// GetDomain retrieves a domain from the repository
func (r *InMemoryDomainRepository) GetDomain(domainName string) (Domain, error) {
	const op errors.Op = "acme/inMemoryDomainRepository.GetDomain"

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.domains == nil {
		return Domain{}, errors.New(op, errors.NotFound, fmt.Sprintf("domain name: %s", domainName))
	}
	if domain, ok := r.domains[domainName]; ok {
		return domain, nil
	}
	return Domain{}, errors.New(op, errors.NotFound, fmt.Sprintf("domain name: %s", domainName))
}

// InMemoryUserRepository is a simple in-memory implementation of
// UserRepository.
type InMemoryUserRepository struct {
	users map[uuid.UUID]User
	mu    sync.Mutex
}

// UpdateUser saves or updates a user in the InMemoryUserRepository using
// the update function f.
//
// UpdateUser passes a pointer to a user object to f. If f does not return
// an error, UpdateUser stores the updated user.
//
// Callers must not hold on to the *User parameter passed to f. Rather
// they should use the user returned by UpdateUser. If UpdateUser returns
// an error, the result of f is not stored in the repository.
func (r *InMemoryUserRepository) UpdateUser(id uuid.UUID, f func(*User) error) (User, error) {
	const op errors.Op = "acme/inMemoryUserRepository.UpdateUser"

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.users == nil {
		r.users = make(map[uuid.UUID]User)
	}
	user := r.users[id]
	err := f(&user)
	if err != nil {
		return User{}, errors.New(op, fmt.Sprintf("update user: %v", id), err)
	}
	r.users[id] = user
	return user, nil
}

// GetUser retrieves the user from the user repository
func (r *InMemoryUserRepository) GetUser(id uuid.UUID) (User, error) {
	const op errors.Op = "acme/inMemoryUserRepository.GetUser"

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.users == nil {
		return User{}, errors.New(op, errors.NotFound, fmt.Sprintf("user id: %v", id))
	}
	if c, ok := r.users[id]; ok {
		return c, nil
	}
	return User{}, errors.New(op, errors.NotFound, fmt.Sprintf("user id: %v", id))
}

// FileBasedCertificateObtainer reads certificates from files and returns them
// when ObtainCertificate is called.
type FileBasedCertificateObtainer struct {
	T        *testing.T // test using this instance of FakeCA.
	CertFile string     // file containing PEM encoded the certificate returned by ObtainCertificate
	KeyFile  string     // file containing PEM encoded the private key returned by ObtainCertificate
}

// ObtainCertificate reads CertFail and KeyFile and returns their contents.
func (c *FileBasedCertificateObtainer) ObtainCertificate(
	req CertificateRequest,
) (
	*CertificateInfo,
	error,
) {
	cert := c.readCertFile()
	key, err := ioutil.ReadFile(c.KeyFile)
	if err != nil {
		c.T.Fatalf("read test key: %v", err)
	}
	certInfo := &CertificateInfo{
		Certificate: cert,
		PrivateKey:  key,
	}
	return certInfo, nil
}

// AssertIssuedCertificate asserts that the passed certificate was issued by
// the file based certificate authority.
func (c *FileBasedCertificateObtainer) AssertIssuedCertificate(cert []byte) {
	expected := c.readCertFile()
	if !reflect.DeepEqual(expected, cert) {
		c.T.Error("certificates do not match")
	}
}

func (c *FileBasedCertificateObtainer) readCertFile() []byte {
	cert, err := ioutil.ReadFile(c.CertFile)
	if err != nil {
		c.T.Fatalf("read test certificate: %v", err)
	}
	return cert
}
