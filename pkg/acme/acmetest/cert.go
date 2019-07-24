package acmetest

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertCertificateValid asserts that the certificate was signed by
// using the issuerCerts for the domain.
func AssertCertificateValid(t *testing.T, domain string, issuerCerts, certificate []byte) {
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(issuerCerts)
	cert := parseCertificate(t, certificate)
	opts := x509.VerifyOptions{
		DNSName: domain,
		Roots:   roots,
	}
	if _, err := cert.Verify(opts); err != nil {
		t.Errorf("Certificate was not valid: %v", err)
	}
}

// AssertKeyBelongsToCertificate asserts that the key belongs to the certificate.
func AssertKeyBelongsToCertificate(t *testing.T, certificate, key []byte) {
	cert := parseCertificate(t, certificate)
	// TODO Is is possible for ACME to issue other key types than RSA?
	publicKey := cert.PublicKey.(*rsa.PublicKey)
	privateKey := parseRSAPrivateKey(t, key)
	assert.Equal(t, publicKey, privateKey.Public())
}

func parseCertificate(t *testing.T, certificate []byte) *x509.Certificate {
	block, _ := pem.Decode(certificate)
	if block == nil {
		t.Fatal("Passed certificate was not PEM encoded")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

func parseRSAPrivateKey(t *testing.T, key []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(key)
	if block == nil {
		t.Fatal("Passed key was not PEM encoded")
	}
	k, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return k
}
