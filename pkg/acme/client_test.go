package acme_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmetest"
	"github.com/fhofherr/acmeproxy/pkg/acme/internal/challenge"
	"github.com/stretchr/testify/assert"
)

const challengeServerPort = 5002

func TestObtainCertificate(t *testing.T) {
	fx, tearDown := newClientTestFixture(t)
	defer tearDown()

	domain := "www.example.com"
	certReq := acme.CertificateRequest{
		Email:         "john.doe@example.com",
		Domains:       []string{domain},
		Bundle:        true,
		CreateAccount: true,
		Key:           newPrivateKey(t),
	}
	certResp, err := fx.Client.ObtainCertificate(certReq)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, certResp.URL)
		acmetest.AssertCertificateValid(t, domain, certResp.IssuerCertificate, certResp.Certificate)
		fx.Pebble.AssertIssuedByPebble(t, domain, certResp.Certificate)
	}
}

func newPrivateKey(t *testing.T) crypto.PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

type clientTestFixture struct {
	Pebble *acmetest.Pebble
	Client acme.Client
}

func newClientTestFixture(t *testing.T) (clientTestFixture, func()) {
	pebble := acmetest.NewPebble(t)
	resetCACerts := acmetest.SetLegoCACertificates(t, pebble.TestCert)
	client := acme.Client{
		DirectoryURL: pebble.DirectoryURL(),
		HTTP01Solver: challenge.NewHTTP01Solver(),
	}
	server := acmetest.NewChallengeServer(t, client.HTTP01Solver, challengeServerPort)
	fixture := clientTestFixture{
		Pebble: pebble,
		Client: client,
	}
	return fixture, func() {
		server.Close()
		resetCACerts()
	}
}
