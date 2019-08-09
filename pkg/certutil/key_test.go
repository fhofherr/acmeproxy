package certutil_test

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
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
		{"RSA2048", certutil.RSA2048, (*rsa.PrivateKey)(nil)},
		{"RSA4096", certutil.RSA4096, (*rsa.PrivateKey)(nil)},
		{"RSA8192", certutil.RSA8192, (*rsa.PrivateKey)(nil)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := certutil.NewPrivateKey(tt.keyType)
			assert.NoError(t, err)
			assert.IsType(t, tt.expectedType, actual)
		})
	}
}

func TestNewPrivateKeyInvalidKeyType(t *testing.T) {
	keyType := certutil.KeyType(-1)
	_, err := certutil.NewPrivateKey(keyType)
	assert.Error(t, err)
}
