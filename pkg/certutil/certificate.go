package certutil

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// ReadCertificate reads an x509 certificate from the passed reader. If
// pemDecode is true ReadCertificateFromFile attempts to PEM decode the file
// before parsing the certificate.
func ReadCertificate(r io.Reader, pemDecode bool) (*x509.Certificate, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "read certificate")
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
	cert, err := x509.ParseCertificate(bs)
	return cert, errors.Wrap(err, "parse certificate")
}

// ReadCertificateFromFile reads an x509 certificate from the passed file. If
// pemDecode is true ReadCertificateFromFile attempts to PEM decode the file
// before parsing the certificate.
func ReadCertificateFromFile(path string, pemDecode bool) (*x509.Certificate, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open certificate file")
	}
	defer r.Close()
	cert, err := ReadCertificate(r, pemDecode)
	if err != nil {
		return nil, errors.Wrap(err, "read certificate file")
	}
	return cert, nil
}

// WriteCertificate writes cert to w. If pemEncode is true the certificate
// is PEM encoded before writing. Otherwise the certificate is written in
// ASN.1 DER encoded form.
func WriteCertificate(cert *x509.Certificate, w io.Writer, pemEncode bool) error {
	var err error
	bs := cert.Raw
	if pemEncode {
		bs, err = pemEncodeBytes("CERTIFICATE", bs)
		if err != nil {
			return errors.Wrap(err, "pem encode certificate")
		}
	}
	_, err = w.Write(bs)
	return errors.Wrap(err, "write certificate")
}

// WriteCertificateToFile writes the passed certificate to the file specified
// by path. If pemEncode is true the certificate is PEM encoded before writing.
// Otherwise the certificate is written in ASN.1 DER encoded form.
func WriteCertificateToFile(cert *x509.Certificate, path string, pemEncode bool) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.Wrap(err, "create directories")
	}
	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "create certificate file")
	}
	defer w.Close()
	return WriteCertificate(cert, w, pemEncode)
}
