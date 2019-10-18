package acme_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegisterNewUser(t *testing.T) {
	tests := []struct {
		name   string
		email  string
		userID uuid.UUID
	}{
		{
			name:   "register new user without E-Mail",
			userID: uuid.Must(uuid.NewRandom()),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fx := newAgentFixture(t, "www.example.com")
			err := fx.Agent.RegisterUser(tt.userID, tt.email)
			assert.NoError(t, err)
			user, err := fx.UserRepository.GetUser(tt.userID)
			assert.NoError(t, err)
			fx.AccountCreator.AssertCreated(t, tt.email, user)
		})
	}
}

func TestRegisterNewDomain(t *testing.T) {
	domainName := "www.example.com"
	fx := newAgentFixture(t, domainName)

	userID := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterUser(userID, "")
	assert.NoError(t, err)

	err = fx.Agent.RegisterDomain(userID, domainName)
	assert.NoError(t, err)

	domain, err := fx.DomainRepository.GetDomain(domainName)
	assert.NoError(t, err)
	assert.Equal(t, domainName, domain.Name)
	assert.Equal(t, userID, domain.UserID)

	// Re-registering the domain again with the same userID must not lead
	// to an error.
	err = fx.Agent.RegisterDomain(userID, domainName)
	assert.NoError(t, err)

	certBuf := bytes.Buffer{}
	err = fx.Agent.WriteCertificate(userID, domainName, &certBuf)
	assert.NoError(t, err)
	fx.FakeCA.AssertIssuedCertificate(certBuf.Bytes())

	keyBuf := bytes.Buffer{}
	err = fx.Agent.WritePrivateKey(userID, domainName, &keyBuf)
	assert.NoError(t, err)
	certutil.AssertKeyBelongsToCertificate(t, acme.DefaultKeyType, certBuf.Bytes(), keyBuf.Bytes())
}

func TestRegisterDomainForUnknownUser(t *testing.T) {
	domainName := "www.example.com"
	fx := newAgentFixture(t, domainName)

	userID := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterDomain(userID, domainName)
	assert.Error(t, err)
}

func TestRegisterSameDomainForDifferentUsers(t *testing.T) {
	domain := "www.example.org"
	fx := newAgentFixture(t, domain)

	userID1 := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterUser(userID1, "")
	assert.NoError(t, err)

	userID2 := uuid.Must(uuid.NewRandom())
	err = fx.Agent.RegisterUser(userID2, "")
	assert.NoError(t, err)

	err = fx.Agent.RegisterDomain(userID1, domain)
	assert.NoError(t, err)

	err = fx.Agent.RegisterDomain(userID2, domain)
	assert.Error(t, err)
}

type agentFixture struct {
	FakeCA           *acme.FileBasedCertificateObtainer
	UserRepository   *acme.InMemoryUserRepository
	DomainRepository *acme.InMemoryDomainRepository
	AccountCreator   *acme.InMemoryAccountCreator
	Agent            *acme.Agent
}

func newAgentFixture(t *testing.T, commonName string) agentFixture {
	keyFile := filepath.Join("testdata", t.Name(), "rsa2048.pem")
	certFile := filepath.Join("testdata", t.Name(), "certificate.pem")
	if *testsupport.FlagUpdate {
		certutil.CreateOpenSSLPrivateKey(t, certutil.RSA2048, keyFile, true)
		certutil.CreateOpenSSLSelfSignedCertificate(t, commonName, keyFile, certFile, true)
	}
	fakeCA := &acme.FileBasedCertificateObtainer{
		CertFile: certFile,
		KeyFile:  keyFile,
		T:        t,
	}
	userRepository := &acme.InMemoryUserRepository{}
	domainRepository := &acme.InMemoryDomainRepository{}
	accountCreator := &acme.InMemoryAccountCreator{}
	agent := &acme.Agent{
		Domains:      domainRepository,
		Users:        userRepository,
		Certificates: fakeCA,
		ACMEAccounts: accountCreator,
	}
	return agentFixture{
		FakeCA:           fakeCA,
		UserRepository:   userRepository,
		DomainRepository: domainRepository,
		AccountCreator:   accountCreator,
		Agent:            agent,
	}
}
