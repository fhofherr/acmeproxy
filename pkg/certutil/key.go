package certutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"

	"github.com/pkg/errors"
)

// KeyType represents the types of cryptographic keys supported by acmeproxy.
//
// The supported key types are dictated by what our ACME client library
// supports.
type KeyType int

const (
	// EC256 represents an ECDSA key using an elliptic curve implementing P-256.
	EC256 KeyType = iota
	// EC384 represents an ECDSA key using an elliptic curve implementing P-384.
	EC384
	// RSA2048 represents an RSA key with a size of 2048 bits.
	RSA2048
	// RSA4096 represents an RSA key with a size of 4096 bits.
	RSA4096
	// RSA8192 represents an RSA key with a size of 8192 bits.
	RSA8192
)

// NewPrivateKey creates a new private key for the specified key type.
//
// It uses crypto/rand.Reader as the source for cryptographically secure
// random numbers.
func NewPrivateKey(kt KeyType) (crypto.PrivateKey, error) {
	var (
		pk  crypto.PrivateKey
		err error
	)
	switch kt {
	case EC256:
		pk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case EC384:
		pk, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case RSA2048:
		pk, err = rsa.GenerateKey(rand.Reader, 2048)
	case RSA4096:
		pk, err = rsa.GenerateKey(rand.Reader, 4096)
	case RSA8192:
		pk, err = rsa.GenerateKey(rand.Reader, 8192)
	default:
		return nil, errors.Errorf("unknown key type: %v", kt)
	}
	return pk, errors.Wrap(err, "new private key")
}
