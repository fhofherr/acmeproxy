package acme_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/acmetest"
	"github.com/stretchr/testify/assert"
)

func TestObtainCertificate(t *testing.T) {
	pebble := acmetest.NewPebble(t)
	reset := acmetest.SetLegoCACertificates(t, pebble.TestCert)
	defer reset()

	acmeClient := acme.Client{
		DirectoryURL:     pebble.DirectoryURL(),
		ChallengeHandler: acme.NewHTTP01Handler(),
	}
	server := acmetest.NewChallengeServer(t, acmeClient.ChallengeHandler, 5002)
	defer server.Close()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	domain := "www.example.com"
	certReq := acme.CertificateRequest{
		Email:         "john.doe@example.com",
		Domains:       []string{domain},
		Bundle:        true,
		CreateAccount: true,
		Key:           key,
	}
	certResp, err := acmeClient.ObtainCertificate(certReq)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, certResp.URL)
		acmetest.AssertCertificateValid(t, domain, certResp.IssuerCertificate, certResp.Certificate)
		pebble.AssertIssuedByPebble(t, domain, certResp.Certificate)
	}
}
