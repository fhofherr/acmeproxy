package acmetest

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
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
func AssertKeyBelongsToCertificate(t *testing.T, kt acmeclient.KeyType, certificate, key []byte) {
	if kt == "" {
		kt = acmeclient.DefaultKeyType
	}
	cert := parseCertificate(t, certificate)
	privateKey := parseSigner(t, kt, key)
	assert.Equal(t, cert.PublicKey, privateKey.Public())
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

func parseSigner(t *testing.T, kt acmeclient.KeyType, key []byte) crypto.Signer {
	var (
		signer crypto.Signer
		err    error
	)

	block, _ := pem.Decode(key)
	if block == nil {
		t.Fatal("Passed key was not PEM encoded")
	}
	switch kt {
	case acmeclient.RSA2048, acmeclient.RSA4096, acmeclient.RSA8192:
		signer, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case acmeclient.EC256, acmeclient.EC384:
		signer, err = x509.ParseECPrivateKey(block.Bytes)
	default:
		t.Fatalf("Unsupported key type: %v", kt)
	}
	if err != nil {
		t.Fatal(err)
	}
	return signer
}
