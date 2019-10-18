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

func TestRegisterNewClient(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		clientID uuid.UUID
	}{
		{
			name:     "register new client without E-Mail",
			clientID: uuid.Must(uuid.NewRandom()),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fx := newAgentFixture(t, "www.example.com")
			err := fx.Agent.RegisterClient(tt.clientID, tt.email)
			assert.NoError(t, err)
			client, err := fx.ClientRepository.GetClient(tt.clientID)
			assert.NoError(t, err)
			fx.AccountCreator.AssertCreated(t, tt.email, client)
		})
	}
}

func TestRegisterNewDomain(t *testing.T) {
	domainName := "www.example.com"
	fx := newAgentFixture(t, domainName)

	clientID := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterClient(clientID, "")
	assert.NoError(t, err)

	err = fx.Agent.RegisterDomain(clientID, domainName)
	assert.NoError(t, err)

	domain, err := fx.DomainRepository.GetDomain(domainName)
	assert.NoError(t, err)
	assert.Equal(t, domainName, domain.Name)
	assert.Equal(t, clientID, domain.ClientID)

	// Re-registering the domain again with the same clientID must not lead
	// to an error.
	err = fx.Agent.RegisterDomain(clientID, domainName)
	assert.NoError(t, err)

	certBuf := bytes.Buffer{}
	err = fx.Agent.WriteCertificate(clientID, domainName, &certBuf)
	assert.NoError(t, err)
	fx.FakeCA.AssertIssuedCertificate(certBuf.Bytes())

	keyBuf := bytes.Buffer{}
	err = fx.Agent.WritePrivateKey(clientID, domainName, &keyBuf)
	assert.NoError(t, err)
	certutil.AssertKeyBelongsToCertificate(t, acme.DefaultKeyType, certBuf.Bytes(), keyBuf.Bytes())
}

func TestRegisterDomainForUnknownClient(t *testing.T) {
	domainName := "www.example.com"
	fx := newAgentFixture(t, domainName)

	clientID := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterDomain(clientID, domainName)
	assert.Error(t, err)
}

func TestRegisterSameDomainForDifferentClients(t *testing.T) {
	domain := "www.example.org"
	fx := newAgentFixture(t, domain)

	clientID1 := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterClient(clientID1, "")
	assert.NoError(t, err)

	clientID2 := uuid.Must(uuid.NewRandom())
	err = fx.Agent.RegisterClient(clientID2, "")
	assert.NoError(t, err)

	err = fx.Agent.RegisterDomain(clientID1, domain)
	assert.NoError(t, err)

	err = fx.Agent.RegisterDomain(clientID2, domain)
	assert.Error(t, err)
}

type agentFixture struct {
	FakeCA           *acme.FileBasedCertificateObtainer
	ClientRepository *acme.InMemoryClientRepository
	DomainRepository *acme.InMemoryDomainRepository
	AccountCreator   *acme.InMemoryAccountCreator
	Agent            *acme.Agent
}

func newAgentFixture(t *testing.T, commonName string) agentFixture {
	keyFile := filepath.Join("testdata", t.Name(), "rsa2048.pem")
	certFile := filepath.Join("testdata", t.Name(), "certificate.pem")
	if *testsupport.FlagUpdate {
		certutil.CreateOpenSSLPrivateKey(t, keyFile)
		certutil.CreateOpenSSLSelfSignedCertificate(t, commonName, keyFile, certFile, true)
	}
	fakeCA := &acme.FileBasedCertificateObtainer{
		CertFile: certFile,
		KeyFile:  keyFile,
		T:        t,
	}
	clientRepository := &acme.InMemoryClientRepository{}
	domainRepository := &acme.InMemoryDomainRepository{}
	accountCreator := &acme.InMemoryAccountCreator{}
	agent := &acme.Agent{
		Domains:      domainRepository,
		Clients:      clientRepository,
		Certificates: fakeCA,
		ACMEAccounts: accountCreator,
	}
	return agentFixture{
		FakeCA:           fakeCA,
		ClientRepository: clientRepository,
		DomainRepository: domainRepository,
		AccountCreator:   accountCreator,
		Agent:            agent,
	}
}
