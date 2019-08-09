package acmeclient

import (
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/go-acme/lego/certcrypto"
	"github.com/pkg/errors"
)

func legoKeyType(kt certutil.KeyType) (certcrypto.KeyType, error) {
	switch kt {
	case certutil.EC256:
		return certcrypto.EC256, nil
	case certutil.EC384:
		return certcrypto.EC384, nil
	case certutil.RSA2048:
		return certcrypto.RSA2048, nil
	case certutil.RSA4096:
		return certcrypto.RSA4096, nil
	case certutil.RSA8192:
		return certcrypto.RSA8192, nil
	default:
		return "", errors.Errorf("unsupported key type: %v", kt)
	}
}
