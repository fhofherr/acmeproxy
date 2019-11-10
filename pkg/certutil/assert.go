package certutil

import (
	"crypto"
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
	cert, err := ParseCertificate(certificate, true)
	if err != nil {
		t.Fatal(err)
	}
	opts := x509.VerifyOptions{
		DNSName: domain,
		Roots:   roots,
	}
	if _, err := cert.Verify(opts); err != nil {
		t.Errorf("Certificate was not valid: %v", err)
	}
}

// AssertKeyBelongsToCertificate asserts that the key belongs to the certificate.
func AssertKeyBelongsToCertificate(t *testing.T, kt KeyType, certificate, key []byte) {
	cert, err := ParseCertificate(certificate, true)
	if !assert.NoError(t, err) {
		return
	}
	privateKey := parseSigner(t, kt, key)
	assert.Equal(t, cert.PublicKey, privateKey.Public(),
		"public key mismatch: key did not belong to certificate.")
}

func parseSigner(t *testing.T, kt KeyType, key []byte) crypto.Signer {
	var (
		signer crypto.Signer
		err    error
	)

	block, _ := pem.Decode(key)
	if block == nil {
		t.Fatal("Passed key was not PEM encoded")
	}
	switch kt {
	case RSA2048, RSA4096, RSA8192:
		signer, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case EC256, EC384, EC521:
		signer, err = x509.ParseECPrivateKey(block.Bytes)
	default:
		t.Fatalf("Unsupported key type: %v", kt)
	}
	if err != nil {
		t.Fatal(err)
	}
	return signer
}
