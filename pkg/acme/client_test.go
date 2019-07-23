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

	acmeClient := acme.Client{DirectoryURL: pebble.DirectoryURL()}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certReq := acme.CertificateRequest{
		Email:         "john.doe@example.com",
		Domains:       []string{"www.example.com"},
		Bundle:        true,
		CreateAccount: true,
		Key:           key,
	}
	certResp, err := acmeClient.ObtainCertificate(certReq)
	if assert.NoError(t, err) {
		// TODO test if certificate is valid
		assert.NotEmpty(t, certResp.Certificate)
	}
}
