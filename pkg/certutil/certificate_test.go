package certutil_test

import (
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestReadCertificate(t *testing.T) {
	tests := []struct {
		name      string
		keyType   certutil.KeyType
		keyFile   string
		certFile  string
		pemEncode bool
	}{
		{
			name:      "read PEM encoded certificate",
			keyType:   certutil.RSA2048,
			keyFile:   "rsa2048.pem",
			certFile:  "certificate.pem",
			pemEncode: true,
		},
		{
			name:      "read ASN.1 DER encoded certificate",
			keyType:   certutil.RSA2048,
			keyFile:   "rsa2048.pem",
			certFile:  "certificate.der",
			pemEncode: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			commonName := "example.com"
			certFile := filepath.Join("testdata", t.Name(), tt.certFile)
			if *testsupport.FlagUpdate {
				keyFile := filepath.Join("testdata", t.Name(), tt.keyFile)
				certutil.CreateOpenSSLPrivateKey(t, tt.keyType, keyFile, true)
				certutil.CreateOpenSSLSelfSignedCertificate(t, commonName, keyFile, certFile, tt.pemEncode)
			}
			_, err := certutil.ReadCertificateFromFile(certFile, tt.pemEncode)
			assert.NoError(t, err)
		})
	}
}

func TestWriteCertificate(t *testing.T) {
	tests := []struct {
		name      string
		keyType   certutil.KeyType
		keyFile   string
		certFile  string
		pemEncode bool
	}{
		{
			name:      "write PEM encoded certificate",
			keyType:   certutil.RSA2048,
			keyFile:   "rsa2048.pem",
			certFile:  "certificate.pem",
			pemEncode: true,
		},
		{
			name:      "write ASN.1 DER encoded certificate",
			keyType:   certutil.RSA2048,
			keyFile:   "rsa2048.pem",
			certFile:  "certificate.der",
			pemEncode: false,
		},
	}
	tmpDir, tearDown := createTmpDir(t)
	defer tearDown()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			commonName := "example.com"
			certFile := filepath.Join("testdata", t.Name(), tt.certFile)
			if *testsupport.FlagUpdate {
				keyFile := filepath.Join("testdata", t.Name(), tt.keyFile)
				certutil.CreateOpenSSLPrivateKey(t, tt.keyType, keyFile, true)
				certutil.CreateOpenSSLSelfSignedCertificate(t, commonName, keyFile, certFile, tt.pemEncode)
			}
			cert, err := certutil.ReadCertificateFromFile(certFile, tt.pemEncode)
			if !assert.NoError(t, err) {
				return
			}
			targetFile := filepath.Join(tmpDir, t.Name(), "cert_file")
			err = certutil.WriteCertificateToFile(cert, targetFile, tt.pemEncode)
			assert.NoError(t, err)
			assert.Equal(t, sha256SumFile(t, certFile), sha256SumFile(t, targetFile))
		})
	}
}
