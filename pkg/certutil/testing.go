package certutil

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	mrand "math/rand"
	"testing"
	"time"
)

// WritePrivateKeyForTesting generates a private key of type kt and writes it
// to keyFile. If pemEncode is true the key is PEM encoded.
func WritePrivateKeyForTesting(t *testing.T, keyFile string, kt KeyType, pemEncode bool) crypto.PrivateKey {
	pk, err := NewPrivateKey(kt)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	err = WritePrivateKeyToFile(pk, keyFile, pemEncode)
	if err != nil {
		t.Fatalf("write private key: %v", err)
	}
	return pk
}

// CreateSelfSignedCertificate uses pk to create a self-signed x509 certificate.
func CreateSelfSignedCertificate(t *testing.T, cn string, pk crypto.PrivateKey) *x509.Certificate {
	key, ok := pk.(crypto.Signer)
	if !ok {
		t.Fatal("pk was not an instance of crypto.Signer")
	}
	serial := big.NewInt(mrand.Int63())
	template := &x509.Certificate{
		DNSNames: []string{cn},
		Subject: pkix.Name{
			CommonName: cn,
		},
		SerialNumber: serial,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(100, 0, 0),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, key.Public(), key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

// WriteCertificateForTesting creates and writes a self-signed certificate for
// use during unit tests. See CreateSelfSignedCertificate for details about
// how the certificate is created.
func WriteCertificateForTesting(
	t *testing.T, certFile string, cn string, pk crypto.PrivateKey, pemEncode bool,
) *x509.Certificate {
	cert := CreateSelfSignedCertificate(t, cn, pk)
	err := WriteCertificateToFile(cert, certFile, pemEncode)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}
