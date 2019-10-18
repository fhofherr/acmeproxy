package certutil

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fhofherr/acmeproxy/pkg/errors"
)

// ParseCertificate reads an x509 certificate. If pemDecode is true
// ParseCertificate attempts to PEM decode the data before parsing the
// certificate.
func ParseCertificate(certificate []byte, pemDecode bool) (*x509.Certificate, error) {
	const op errors.Op = "certutil/parseCertificate"

	if pemDecode {
		block, _ := pem.Decode(certificate)
		if block == nil {
			return nil, errors.New(op, "pem decode")
		}
		certificate = block.Bytes
	}
	cert, err := x509.ParseCertificate(certificate)
	if err != nil {
		return nil, errors.New(op, "parse certificate", err)
	}
	return cert, nil
}

// ReadCertificate reads an x509 certificate from the passed reader. If
// pemDecode is true ReadCertificateFromFile attempts to PEM decode the file
// before parsing the certificate.
func ReadCertificate(r io.Reader, pemDecode bool) (*x509.Certificate, error) {
	const op errors.Op = "certutil/readCertificate"

	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.New(op, "read certificate", err)
	}
	return ParseCertificate(bs, pemDecode)
}

// ReadCertificateFromFile reads an x509 certificate from the passed file. If
// pemDecode is true ReadCertificateFromFile attempts to PEM decode the file
// before parsing the certificate.
func ReadCertificateFromFile(path string, pemDecode bool) (*x509.Certificate, error) {
	const op errors.Op = "certutil/readCertificateFromFile"

	r, err := os.Open(path)
	if err != nil {
		return nil, errors.New(op, "open certificate file", err)
	}
	defer r.Close()

	return ReadCertificate(r, pemDecode)
}

// WriteCertificate writes cert to w. If pemEncode is true the certificate
// is PEM encoded before writing. Otherwise the certificate is written in
// ASN.1 DER encoded form.
func WriteCertificate(cert *x509.Certificate, w io.Writer, pemEncode bool) error {
	const op errors.Op = "certutil/writeCertificate"
	var err error

	bs := cert.Raw
	if pemEncode {
		bs, err = pemEncodeBytes("CERTIFICATE", bs)
		if err != nil {
			return errors.New(op, "pem encode certificate", err)
		}
	}
	_, err = w.Write(bs)
	return errors.Wrap(err, op, "write certificate")
}

// WriteCertificateToFile writes the passed certificate to the file specified
// by path. If pemEncode is true the certificate is PEM encoded before writing.
// Otherwise the certificate is written in ASN.1 DER encoded form.
func WriteCertificateToFile(cert *x509.Certificate, path string, pemEncode bool) error {
	const op errors.Op = "certutil/writeCertificateToFile"

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.New(op, "create directories", err)
	}

	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New(op, "open certificate file", err)
	}
	defer w.Close()

	return WriteCertificate(cert, w, pemEncode)
}
