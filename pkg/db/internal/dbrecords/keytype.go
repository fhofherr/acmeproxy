package dbrecords

import (
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/errors"
)

type keyType uint32

const (
	ecdsa keyType = iota
	rsa
)

func keyTypeFromCertutil(kt certutil.KeyType) (keyType, error) {
	const op errors.Op = "dbrecords/keyTypeFromCertutil"

	switch kt {
	case certutil.EC256, certutil.EC384:
		return ecdsa, nil
	case certutil.RSA2048, certutil.RSA4096, certutil.RSA8192:
		return rsa, nil
	default:
		return 0, errors.New(op, "unsupported key type")
	}
}
