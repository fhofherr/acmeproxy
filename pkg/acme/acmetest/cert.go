package acmetest

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// AssertCertificateValid asserts that the certificate was signed by
// using the issuerCerts for the domain.
func AssertCertificateValid(t *testing.T, domain string, issuerCerts, certificate []byte) {
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(issuerCerts)
	block, _ := pem.Decode(certificate)
	if block == nil {
		t.Fatal("Passed issuer certificates were not PEM encoded")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
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
