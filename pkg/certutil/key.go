package certutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"

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

// ReadPrivateKey reads an private key from r using either ReadECDSAPrivateKey
// or ReadRSAPrivateKey.
//
// The value of kt determines which the type of key to be read. To read an
// ECDSA private key any of the EC* values can be used. Likewise to read an RSA
// private key any of the RSA* values can be passed.
func ReadPrivateKey(kt KeyType, r io.Reader, pemDecode bool) (crypto.PrivateKey, error) {
	var (
		pk  crypto.PrivateKey
		err error
	)
	switch kt {
	case EC256, EC384:
		pk, err = readKey(r, pemDecode, parseECDSAKey)
	case RSA2048, RSA4096, RSA8192:
		pk, err = readKey(r, pemDecode, parseRSAKey)
	default:
		return nil, errors.Errorf("invalid key type: %v", kt)
	}
	return pk, errors.Wrap(err, "read private key")
}

func readKey(r io.Reader, pemDecode bool, df func([]byte) (crypto.PrivateKey, error)) (crypto.PrivateKey, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "read key data")
	}
	if pemDecode {
		block, rest := pem.Decode(bs)
		if block == nil {
			return nil, errors.New("empty PEM block")
		}
		if len(rest) > 0 {
			return nil, errors.New("found more than one PEM block")
		}
		bs = block.Bytes
	}
	return df(bs)
}

func parseECDSAKey(bs []byte) (crypto.PrivateKey, error) {
	key, err := x509.ParseECPrivateKey(bs)
	return key, errors.Wrap(err, "parse ECDSA private key")
}

func parseRSAKey(bs []byte) (crypto.PrivateKey, error) {
	key, err := x509.ParsePKCS1PrivateKey(bs)
	return key, errors.Wrap(err, "parse RSA private key")
}
