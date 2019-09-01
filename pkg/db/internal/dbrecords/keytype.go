package dbrecords

import (
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/pkg/errors"
)

type keyType uint32

const (
	ecdsa keyType = iota
	rsa
)

func keyTypeFromCertutil(kt certutil.KeyType) (keyType, error) {
	switch kt {
	case certutil.EC256, certutil.EC384:
		return ecdsa, nil
	case certutil.RSA2048, certutil.RSA4096, certutil.RSA8192:
		return rsa, nil
	default:
		return 0, errors.New("unsupported certutil key type")
	}
}
