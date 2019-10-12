package acmetest

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
)

// FakeCA implements a fake certificate authority which issues certificates.
//
// It implements the acme.CertificateObtainer interface and can thus be used
// to test the acme.Agent without needing a real ACME CA.
//
// FakeCA is heavily inspired by minica: https://github.com/jsha/minica
//
// Deprecated: this is way to much code to maintain just for testing.
type FakeCA struct {
	T        *testing.T        // test using this instance of FakeCA.
	KeyBits  int               // Bit size of key.
	rootKey  *rsa.PrivateKey   // private key for the rootCert
	rootCert *x509.Certificate // certificate used to sign the issued certificates
	serial   int64             // serial numbers of certificates
	mu       sync.Mutex        // protect access to internal state
	once     sync.Once
}

// ObtainCertificate creates a certificate for the domains in the request.
func (c *FakeCA) ObtainCertificate(req acmeclient.CertificateRequest) (*acmeclient.CertificateInfo, error) {
	c.initialize()

	cert, key := c.sign(req.Domains)
	certInfo := &acmeclient.CertificateInfo{
		Certificate:       pemEncode(c.T, "CERTIFICATE", cert.Raw),
		PrivateKey:        pemEncode(c.T, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)),
		IssuerCertificate: pemEncode(c.T, "CERTIFICATE", c.rootCert.Raw),
	}
	return certInfo, nil
}

// AssertIssued checks if the certificate has been issued by this FakeCA.
func (c *FakeCA) AssertIssued(t *testing.T, domainName string, cert []byte) {
	c.initialize()

	c.mu.Lock()
	defer c.mu.Unlock()
	rootCert := pemEncode(t, "CERTIFICATE", c.rootCert.Raw)
	certutil.AssertCertificateValid(t, domainName, rootCert, cert)
}

func (c *FakeCA) sign(domains []string) (*x509.Certificate, *rsa.PrivateKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(domains) < 1 {
		c.T.Fatal("at least one domain required")
	}
	key := makeKey(c.T, c.KeyBits)
	commonName := domains[0]
	serial := big.NewInt(c.serial)
	c.serial++
	template := &x509.Certificate{
		DNSNames: domains,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		SerialNumber: serial,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, 1),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, c.rootCert, key.Public(), c.rootKey)
	if err != nil {
		c.T.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		c.T.Fatal(err)
	}
	return cert, key
}

func (c *FakeCA) initialize() {
	c.once.Do(func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.KeyBits == 0 {
			c.KeyBits = 2048
		}
		c.rootKey = makeKey(c.T, c.KeyBits)
		c.rootCert = makeRootCert(c.T, c.rootKey, big.NewInt(c.serial))
		c.serial++
	})
}

func makeKey(t *testing.T, keyBits int) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func makeRootCert(t *testing.T, key *rsa.PrivateKey, serial *big.Int) *x509.Certificate {
	skid := calculateSKID(t, key.Public())
	template := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "acmeproxy FakeCA",
		},
		SerialNumber: serial,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, 1),

		SubjectKeyId:          skid,
		AuthorityKeyId:        skid,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
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

func calculateSKID(t *testing.T, pubKey crypto.PublicKey) []byte {
	spkiASN1, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		t.Fatal(err)
	}
	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	if err != nil {
		t.Fatal(err)
	}
	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)
	return skid[:]
}

func pemEncode(t *testing.T, typ string, der []byte) []byte {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  typ,
		Bytes: der,
	})
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
