package acmeclient

import (
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/go-acme/lego/certcrypto"
	"github.com/stretchr/testify/assert"
)

func TestKeyTypeToLegoKeyType(t *testing.T) {
	tests := []struct {
		name        string
		keyType     certutil.KeyType
		legoKeyType certcrypto.KeyType
	}{
		{"EC256", certutil.EC256, certcrypto.EC256},
		{"EC384", certutil.EC384, certcrypto.EC384},
		{"RSA2048", certutil.RSA2048, certcrypto.RSA2048},
		{"RSA4096", certutil.RSA4096, certcrypto.RSA4096},
		{"RSA8192", certutil.RSA8192, certcrypto.RSA8192},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := legoKeyType(tt.keyType)
			assert.NoError(t, err)
			assert.Equal(t, tt.legoKeyType, actual)
		})
	}
}

func TestInvalidKeyTypeToLegoKeyType(t *testing.T) {
	_, err := legoKeyType(certutil.KeyType(-1))
	assert.Error(t, err)
}
