package acmeclient_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmetest"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/stretchr/testify/assert"
)

const challengeServerPort = 5002

func TestCreateAccount(t *testing.T) {
	acmetest.SkipIfPebbleDisabled(t)
	tests := []struct {
		name  string
		email string
	}{
		{name: "create account without email"},
		{name: "create account with email", email: "jane.doe@example.com"},
	}

	fx, tearDown := newClientTestFixture(t)
	defer tearDown()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			accountKey := newPrivateKey(t)
			accountURL, err := fx.Client.CreateAccount(accountKey, tt.email)
			assert.NoError(t, err)
			assert.NotEmpty(t, accountURL)
			assert.Truef(
				t,
				strings.HasPrefix(accountURL, fx.Pebble.AccountURLPrefix()),
				"accountURL %s did not start with %s",
				accountURL,
				fx.Pebble.AccountURLPrefix())
		})
	}

}

func TestObtainCertificate(t *testing.T) {
	acmetest.SkipIfPebbleDisabled(t)

	fx, tearDown := newClientTestFixture(t)
	defer tearDown()

	tests := []struct {
		acmeclient.CertificateRequest
		name string
	}{
		{
			name: "obtain certificate without account",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA2048@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: newPrivateKey(t),
			},
		},
		{
			name: "obtain RSA4096 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA4096@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: newPrivateKey(t),
				KeyType:    certutil.RSA4096,
			},
		},
		{
			name: "obtain RSA8192 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA8192@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: newPrivateKey(t),
				KeyType:    certutil.RSA8192,
			},
		},
		{
			name: "obtain EC256 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+EC256@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: newPrivateKey(t),
				KeyType:    certutil.EC256,
			},
		},
		{
			name: "obtain EC384 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+EC384@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: newPrivateKey(t),
				KeyType:    certutil.EC384,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			certInfo, err := fx.Client.ObtainCertificate(tt.CertificateRequest)
			if !assert.NoError(t, err) {
				return
			}
			assert.NotEmpty(t, certInfo.URL)
			assert.NotEmpty(t, certInfo.AccountURL)
			for _, domain := range tt.CertificateRequest.Domains {
				certutil.AssertCertificateValid(t, domain, certInfo.IssuerCertificate, certInfo.Certificate)
				certutil.AssertKeyBelongsToCertificate(t, tt.KeyType, certInfo.Certificate, certInfo.PrivateKey)
				fx.Pebble.AssertIssuedByPebble(t, domain, certInfo.Certificate)
			}
		})
	}
}

func TestObtainCertificateWithPreExistingAccount(t *testing.T) {
	acmetest.SkipIfPebbleDisabled(t)

	fx, tearDown := newClientTestFixture(t)
	defer tearDown()

	domain := "www.example.com"
	accountKey := newPrivateKey(t)
	accountURL, err := fx.Client.CreateAccount(accountKey, "jane.doe@example.com")
	assert.NoError(t, err)

	req := acmeclient.CertificateRequest{
		Email:      "jane.doe+RSA2048@example.com",
		AccountURL: accountURL,
		Domains:    []string{domain},
		Bundle:     true,
		AccountKey: accountKey,
		KeyType:    certutil.RSA2048,
	}
	ci, err := fx.Client.ObtainCertificate(req)
	assert.NoError(t, err)
	fx.Pebble.AssertIssuedByPebble(t, domain, ci.Certificate)
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
	Client acmeclient.Client
}

func newClientTestFixture(t *testing.T) (clientTestFixture, func()) {
	pebble := acmetest.NewPebble(t)
	resetCACerts := acmetest.SetLegoCACertificates(t, pebble.TestCert)
	client := acmeclient.Client{
		DirectoryURL: pebble.DirectoryURL(),
		HTTP01Solver: acmeclient.NewHTTP01Solver(),
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
