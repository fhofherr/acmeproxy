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
	if acmetest.SkipIfPebbleDisabled(t) {
		return
	}

	fx, tearDown := newClientTestFixture(t)
	defer tearDown()
	privateKey := newPrivateKey(t)

	tests := []struct {
		acme.CertificateRequest
		name string
	}{
		{
			name: "obtain certificate without account",
			CertificateRequest: acme.CertificateRequest{
				Email:      "john.doe+RSA2048@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
			},
		},
		{
			name: "obtain certificate with pre-existing account",
			CertificateRequest: acme.CertificateRequest{
				Email:      "jane.doe+RSA2048@example.com",
				AccountURL: fx.Pebble.CreateAccount(t, "jane.doe@example.com", privateKey),
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acme.RSA2048,
			},
		},
		{
			name: "obtain RSA4096 certificate",
			CertificateRequest: acme.CertificateRequest{
				Email:      "john.doe+RSA4096@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acme.RSA4096,
			},
		},
		{
			name: "obtain RSA8192 certificate",
			CertificateRequest: acme.CertificateRequest{
				Email:      "john.doe+RSA8192@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acme.RSA8192,
			},
		},
		{
			name: "obtain EC256 certificate",
			CertificateRequest: acme.CertificateRequest{
				Email:      "john.doe+EC256@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acme.EC256,
			},
		},
		{
			name: "obtain EC384 certificate",
			CertificateRequest: acme.CertificateRequest{
				Email:      "john.doe+EC384@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acme.EC384,
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
				acmetest.AssertCertificateValid(t, domain, certInfo.IssuerCertificate, certInfo.Certificate)
				acmetest.AssertKeyBelongsToCertificate(t, tt.KeyType, certInfo.Certificate, certInfo.PrivateKey)
				fx.Pebble.AssertIssuedByPebble(t, domain, certInfo.Certificate)
			}
		})
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
