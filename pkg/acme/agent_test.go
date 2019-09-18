package acme_test

import (
	"bytes"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmetest"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
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
			fx := newAgentFixture(t)
			err := fx.Agent.RegisterClient(tt.clientID, tt.email)
			assert.NoError(t, err)
			client, err := fx.ClientRepository.GetClient(tt.clientID)
			assert.NoError(t, err)
			fx.AccountCreator.AssertCreated(t, tt.email, client)
		})
	}

}

func TestRegisterNewDomain(t *testing.T) {
	fx := newAgentFixture(t)

	clientID := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterClient(clientID, "")
	assert.NoError(t, err)

	domainName := "www.example.com"
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
	fx.FakeCA.AssertIssued(t, domainName, certBuf.Bytes())

	keyBuf := bytes.Buffer{}
	err = fx.Agent.WritePrivateKey(clientID, domainName, &keyBuf)
	assert.NoError(t, err)
	certutil.AssertKeyBelongsToCertificate(t, acmeclient.DefaultKeyType, certBuf.Bytes(), keyBuf.Bytes())
}

func TestRegisterDomainForUnknownClient(t *testing.T) {
	fx := newAgentFixture(t)

	clientID := uuid.Must(uuid.NewRandom())
	domainName := "www.example.com"
	err := fx.Agent.RegisterDomain(clientID, domainName)
	assert.Error(t, err)
}

func TestRegisterSameDomainForDifferentClients(t *testing.T) {
	fx := newAgentFixture(t)

	clientID1 := uuid.Must(uuid.NewRandom())
	err := fx.Agent.RegisterClient(clientID1, "")
	assert.NoError(t, err)

	clientID2 := uuid.Must(uuid.NewRandom())
	err = fx.Agent.RegisterClient(clientID2, "")
	assert.NoError(t, err)

	domain := "www.example.com"
	err = fx.Agent.RegisterDomain(clientID1, domain)
	assert.NoError(t, err)

	err = fx.Agent.RegisterDomain(clientID2, domain)
	assert.Error(t, err)
}

type agentFixture struct {
	FakeCA           *acmetest.FakeCA
	ClientRepository *acme.InMemoryClientRepository
	DomainRepository *acme.InMemoryDomainRepository
	AccountCreator   *acme.InMemoryAccountCreator
	Agent            *acme.Agent
}

func newAgentFixture(t *testing.T) agentFixture {
	fakeCA := &acmetest.FakeCA{T: t}
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
