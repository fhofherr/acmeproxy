package certutil

import (
	"crypto"
	"testing"
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
