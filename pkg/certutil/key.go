package certutil

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fhofherr/acmeproxy/pkg/errors"
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
	// EC521 represents an ECDSA key using an elliptic curve implementing P-521.
	EC521
	// RSA2048 represents an RSA key with a size of 2048 bits.
	RSA2048
	// RSA4096 represents an RSA key with a size of 4096 bits.
	RSA4096
	// RSA8192 represents an RSA key with a size of 8192 bits.
	RSA8192
)

// DetermineKeyType inspects the passed key and returns the appropriate
// KeyType. It returns an error if it could not determine the passed key
// type. In this case the returned key type is wrong and has to be ignored.
func DetermineKeyType(key crypto.PrivateKey) (KeyType, error) {
	const op errors.Op = "certutil/DetermineKeyType"

	switch pk := key.(type) {
	case *ecdsa.PrivateKey:
		return determineECDSAKeyType(pk)
	case *rsa.PrivateKey:
		return determineRSAKeyType(pk)
	default:
		return -1, errors.New(op, "unsupported key type")
	}
}

func determineECDSAKeyType(pk *ecdsa.PrivateKey) (KeyType, error) {
	const op errors.Op = "certutil/determineECDSAKeyType"

	curveName := pk.Curve.Params().Name
	switch curveName {
	case "P-256":
		return EC256, nil
	case "P-384":
		return EC384, nil
	case "P-521":
		return EC521, nil
	default:
		return -1, errors.New(op, fmt.Sprintf("unsupported curve: %s", curveName))
	}
}

func determineRSAKeyType(pk *rsa.PrivateKey) (KeyType, error) {
	const op errors.Op = "certutil/determineRSAKeyType"

	bitLen := pk.PublicKey.N.BitLen()
	switch bitLen {
	case 2048:
		return RSA2048, nil
	case 4096:
		return RSA4096, nil
	case 8192:
		return RSA8192, nil
	default:
		return -1, errors.New(op, fmt.Sprintf("unsupported bit length: %d", bitLen))
	}
}

// NewPrivateKey creates a new private key for the specified key type.
//
// It uses crypto/rand.Reader as the source for cryptographically secure
// random numbers.
func NewPrivateKey(kt KeyType) (crypto.PrivateKey, error) {
	const op errors.Op = "certutil/NewPrivateKey"
	var (
		pk  crypto.PrivateKey
		err error
	)
	switch kt {
	case EC256:
		pk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case EC384:
		pk, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case EC521:
		pk, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	case RSA2048:
		pk, err = rsa.GenerateKey(rand.Reader, 2048)
	case RSA4096:
		pk, err = rsa.GenerateKey(rand.Reader, 4096)
	case RSA8192:
		pk, err = rsa.GenerateKey(rand.Reader, 8192)
	default:
		return nil, errors.New(op, fmt.Sprintf("unknown key type: %v", kt))
	}
	return pk, errors.Wrap(err, op, "generate key")
}

// ReadPrivateKey reads an private key from r using either ReadECDSAPrivateKey
// or ReadRSAPrivateKey.
//
// The value of kt determines which the type of key to be read. To read an
// ECDSA private key any of the EC* values can be used. Likewise to read an RSA
// private key any of the RSA* values can be passed.
func ReadPrivateKey(kt KeyType, r io.Reader, pemDecode bool) (crypto.PrivateKey, error) {
	const op errors.Op = "certutil/ReadPrivateKey"

	var (
		pk  crypto.PrivateKey
		err error
	)
	switch kt {
	case EC256, EC384, EC521:
		pk, err = readKey(r, pemDecode, parseECDSAKey)
	case RSA2048, RSA4096, RSA8192:
		pk, err = readKey(r, pemDecode, parseRSAKey)
	default:
		return nil, errors.New(op, fmt.Sprintf("invalid key type: %v", kt))
	}
	return pk, errors.Wrap(err, op)
}

func readKey(r io.Reader, pemDecode bool, df func([]byte) (crypto.PrivateKey, error)) (crypto.PrivateKey, error) {
	const op errors.Op = "certutil/readKey"

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.New(op, "read key data", err)
	}
	if pemDecode {
		block, rest := pem.Decode(bs)
		if block == nil {
			return nil, errors.New(op, "empty PEM block")
		}
		if len(rest) > 0 {
			return nil, errors.New(op, "found more than one PEM block")
		}
		bs = block.Bytes
	}
	return df(bs)
}

func parseECDSAKey(bs []byte) (crypto.PrivateKey, error) {
	const op errors.Op = "certutil/parseECDSAKey"

	key, err := x509.ParseECPrivateKey(bs)
	return key, errors.Wrap(err, op, "parse ECDSA private key")
}

func parseRSAKey(bs []byte) (crypto.PrivateKey, error) {
	const op errors.Op = "certutil/parseRSAKey"

	key, err := x509.ParsePKCS1PrivateKey(bs)
	return key, errors.Wrap(err, op, "parse RSA private key")
}

// ReadPrivateKeyFromFile reads a private key of type kt from the file
// at the specified path. If pemDecode is true ReadPrivateKeyFromFile assumes
// the key is PEM encoded and decodes it accordingly.
func ReadPrivateKeyFromFile(kt KeyType, path string, pemDecode bool) (crypto.PrivateKey, error) {
	const op errors.Op = "certutil/ReadPrivateKeyFromFile"

	keyReader, err := os.Open(path)
	if err != nil {
		return nil, errors.New(op, "open key path", err)
	}
	defer keyReader.Close()
	pk, err := ReadPrivateKey(kt, keyReader, pemDecode)
	if err != nil {
		return nil, errors.New(op, "read key from file", err)
	}
	return pk, nil
}

// WritePrivateKey writes a private key to a file.
//
// WritePrivateKey returns an error if the writing the key to w fails or if
// WritePrivateKey does not support the type of private key passed.
//
// If pemEncode is true WritePrivateKey PEM-encodes the private key before it
// writes it to w.
func WritePrivateKey(key crypto.PrivateKey, w io.Writer, pemEncode bool) error {
	const op errors.Op = "certutil/WritePrivateKey"

	var (
		bs  []byte
		typ string
		err error
	)

	switch pk := key.(type) {
	case *ecdsa.PrivateKey:
		typ = "EC PRIVATE KEY"
		bs, err = x509.MarshalECPrivateKey(pk)
		if err != nil {
			return errors.New(op, "marshal ECDSA private key", err)
		}
	case *rsa.PrivateKey:
		bs = x509.MarshalPKCS1PrivateKey(pk)
		typ = "RSA PRIVATE KEY"
	default:
		return errors.New(op, "unsupported private key")
	}
	if pemEncode {
		bs, err = pemEncodeBytes(typ, bs)
		if err != nil {
			return err
		}
	}
	_, err = w.Write(bs)
	return errors.Wrap(err, op, "write private key")
}

func pemEncodeBytes(typ string, bs []byte) ([]byte, error) {
	const op errors.Op = "certutil/pemEncodeBytes"

	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  typ,
		Bytes: bs,
	})
	return buf.Bytes(), errors.Wrap(err, op, "pem encode")
}

// WritePrivateKeyToFile writes the private key into the file given by path.
//
// If pemEncode is true it will PEM encode the private key before writing it.
//
// WritePrivateKeyToFile creates any missing intermediate directories.
func WritePrivateKeyToFile(key crypto.PrivateKey, path string, pemEncode bool) error {
	const op errors.Op = "certutil/WritePrivateKeyToFile"

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New(op, "create directories", err)
	}
	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New(op, "create key file", err)
	}
	defer w.Close()
	return WritePrivateKey(key, w, pemEncode)
}

// KeyMust panics err != nil. It returns key otherwise.
// KeyMust should not be called in production code unless the caller is
// absolutely sure that a panic is warranted.
func KeyMust(key crypto.PrivateKey, err error) crypto.PrivateKey {
	if err != nil {
		panic(fmt.Sprintf("key must: %v", err))
	}
	return key
}
