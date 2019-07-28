package acmeclient_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmetest"
	"github.com/fhofherr/acmeproxy/pkg/acme/internal/acmeclient"
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
		acmeclient.CertificateRequest
		name string
	}{
		{
			name: "obtain certificate without account",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA2048@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
			},
		},
		{
			name: "obtain certificate with pre-existing account",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "jane.doe+RSA2048@example.com",
				AccountURL: fx.Pebble.CreateAccount(t, "jane.doe@example.com", privateKey),
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acmeclient.RSA2048,
			},
		},
		{
			name: "obtain RSA4096 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA4096@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acmeclient.RSA4096,
			},
		},
		{
			name: "obtain RSA8192 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA8192@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acmeclient.RSA8192,
			},
		},
		{
			name: "obtain EC256 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+EC256@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acmeclient.EC256,
			},
		},
		{
			name: "obtain EC384 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+EC384@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				PrivateKey: privateKey,
				KeyType:    acmeclient.EC384,
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
