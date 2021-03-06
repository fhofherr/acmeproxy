package certutil_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestNewPrivateKey(t *testing.T) {
	// Those tests rely on a cryptographically secure random number generator
	// which is used in certutil.NewPrivateKey. This can be rather slow.
	tests := []struct {
		name         string
		keyType      certutil.KeyType
		expectedType interface{}
	}{
		{"EC256", certutil.EC256, (*ecdsa.PrivateKey)(nil)},
		{"EC384", certutil.EC384, (*ecdsa.PrivateKey)(nil)},
		{"EC521", certutil.EC521, (*ecdsa.PrivateKey)(nil)},
		{"RSA2048", certutil.RSA2048, (*rsa.PrivateKey)(nil)},
		{"RSA4096", certutil.RSA4096, (*rsa.PrivateKey)(nil)},
		{"RSA8192", certutil.RSA8192, (*rsa.PrivateKey)(nil)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := certutil.NewPrivateKey(tt.keyType)
			if !assert.NoError(t, err) {
				return
			}
			keyType, err := certutil.DetermineKeyType(actual)
			if !assert.NoError(t, err) {
				return
			}
			assert.IsType(t, tt.expectedType, actual)
			assert.Equal(t, tt.keyType, keyType)
		})
	}
}

func TestNewPrivateKeyInvalidKeyType(t *testing.T) {
	keyType := certutil.KeyType(-1)
	_, err := certutil.NewPrivateKey(keyType)
	assert.Error(t, err)
}

func TestReadPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		keyType certutil.KeyType
		pem     bool
	}{
		{"ec256.pem", certutil.EC256, true},
		{"ec256.der", certutil.EC256, false},
		{"ec384.pem", certutil.EC384, true},
		{"ec384.der", certutil.EC384, false},
		{"ec521.der", certutil.EC521, false},
		{"ec521.pem", certutil.EC521, true},
		{"rsa2048.pem", certutil.RSA2048, true},
		{"rsa2048.der", certutil.RSA2048, false},
		{"rsa4096.der", certutil.RSA4096, false},
		{"rsa8192.der", certutil.RSA8192, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			keyPath := filepath.Join("testdata", t.Name())
			if *testsupport.FlagUpdate {
				certutil.CreateOpenSSLPrivateKey(t, tt.keyType, keyPath, tt.pem)
			}
			r, err := os.Open(keyPath)
			if !assert.NoError(t, err) {
				return
			}
			_, err = certutil.ReadPrivateKey(tt.keyType, r, tt.pem)
			assert.NoError(t, err)
		})
	}
}

func TestReadPrivateKeyInvalidKeyType(t *testing.T) {
	keyType := certutil.KeyType(-1)
	r := &bytes.Buffer{}
	_, err := certutil.ReadPrivateKey(keyType, r, false)
	assert.Error(t, err)
}

func TestReadPrivateKeyReaderError(t *testing.T) {
	expectedErr := errors.New("expected error")
	r := &errorReader{expectedErr}
	_, err := certutil.ReadPrivateKey(certutil.EC256, r, false)
	assert.Error(t, err)
}

func TestReadPrivateKeyInvalidPEMBlock(t *testing.T) {
	r := strings.NewReader("invalid PEM data")
	_, err := certutil.ReadPrivateKey(certutil.RSA2048, r, true)
	assert.Error(t, err)
}

func TestReadConcatenatedPEMBlocks(t *testing.T) {
	certFiles := []string{
		filepath.Join("testdata", t.Name(), "ec256_1.pem"),
		filepath.Join("testdata", t.Name(), "ec256_2.pem"),
	}
	if *testsupport.FlagUpdate {
		for _, path := range certFiles {
			certutil.CreateOpenSSLPrivateKey(t, certutil.EC256, path, true)
		}
	}
	pemBytes := make([]byte, 0, 1024)
	for _, path := range certFiles {
		bs, err := ioutil.ReadFile(path)
		if !assert.NoError(t, err) {
			return
		}
		pemBytes = append(pemBytes, bs...)
	}
	r := bytes.NewReader(pemBytes)
	_, err := certutil.ReadPrivateKey(certutil.EC256, r, true)
	assert.Error(t, err)
}

func TestWritePrivateKey(t *testing.T) {
	tests := []struct {
		name      string
		keyType   certutil.KeyType
		pemEncode bool
	}{
		{"ec256.pem", certutil.EC256, true},
		{"ec256.der", certutil.EC256, false},
		{"ec384.pem", certutil.EC384, true},
		{"ec384.der", certutil.EC384, false},
		{"ec521.pem", certutil.EC521, true},
		{"ec521.der", certutil.EC521, false},
		{"rsa2048.pem", certutil.RSA2048, true},
		{"rsa2048.der", certutil.RSA2048, false},
		{"rsa4096.der", certutil.RSA4096, false},
		{"rsa8192.der", certutil.RSA8192, false},
	}
	// Create tmpDir before we iterate over our test cases. This way the tmpDir
	// has the test function's name as prefix and does not contain the test
	// cases names.
	tmpDir, tearDown := createTmpDir(t)
	defer tearDown()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srcKeyPath := filepath.Join("testdata", t.Name())
			if *testsupport.FlagUpdate {
				certutil.CreateOpenSSLPrivateKey(t, tt.keyType, srcKeyPath, tt.pemEncode)
			}
			pk := certutil.KeyMust(certutil.ReadPrivateKeyFromFile(tt.keyType, srcKeyPath, tt.pemEncode))
			targetKeyPath := filepath.Join(tmpDir, tt.name)
			w, err := os.Create(targetKeyPath)
			if err != nil {
				t.Fatalf("open target key path: %v", err)
			}
			defer w.Close()
			err = certutil.WritePrivateKey(pk, w, tt.pemEncode)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, sha256SumFile(t, srcKeyPath), sha256SumFile(t, targetKeyPath))
		})
	}
}

type errorReader struct {
	Err error
}

func (e *errorReader) Read([]byte) (int, error) {
	return 0, e.Err
}

func sha256SumFile(t *testing.T, path string) []byte {
	r, err := os.Open(path)
	if err != nil {
		t.Fatalf("open path: %v", err)
	}
	defer r.Close()
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("read bytes from file: %v", err)
	}
	hash := sha256.Sum256(bs)
	return hash[:]
}

func createTmpDir(t *testing.T) (string, func()) {
	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	return tmpDir, func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Error(err)
		}
	}
}
