package acmeclient

import (
	"crypto"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
	"github.com/pkg/errors"
)

// KeyType represents the key algorithm to use.
type KeyType certcrypto.KeyType

// Convenience aliases for all key types supported by lego.
const (
	EC256   = KeyType(certcrypto.EC256)
	EC384   = KeyType(certcrypto.EC384)
	RSA2048 = KeyType(certcrypto.RSA2048)
	RSA4096 = KeyType(certcrypto.RSA4096)
	RSA8192 = KeyType(certcrypto.RSA8192)
)

// DefaultKeyType is the default key type to use if the CertificateRequest does
// not specify one.
const DefaultKeyType = RSA2048

// Client is an ACME protocol client capable of obtaining and renewing
// certificates.
type Client struct {
	DirectoryURL string
	HTTP01Solver *HTTP01Solver
}

// CreateAccount creates a new ACME account for the accountKey.
//
// If email is not empty it is used as the contact address for the new account.
func (c *Client) CreateAccount(accountKey crypto.PrivateKey, email string) (string, error) {
	user := &User{
		Email:      email,
		PrivateKey: accountKey,
	}
	cfg := lego.NewConfig(user)
	cfg.CADirURL = c.DirectoryURL
	legoClient, err := lego.NewClient(cfg)
	if err != nil {
		return "", errors.Wrap(err, "create client for new ACME account")
	}
	err = user.Register(legoClient)
	if err != nil {
		return "", errors.Wrap(err, "register new ACME account")
	}
	return user.Registration.URI, nil
}

// ObtainCertificate obtains a new certificate from the remote ACME server.
func (c *Client) ObtainCertificate(req CertificateRequest) (*CertificateInfo, error) {
	// TODO err if len(req.Domains) < 1
	keyType := certcrypto.KeyType(req.KeyType)
	if keyType == "" {
		keyType = certcrypto.KeyType(DefaultKeyType)
	}
	if req.AccountURL == "" {
		var err error
		req.AccountURL, err = c.CreateAccount(req.AccountKey, req.Email)
		if err != nil {
			return nil, errors.Wrap(err, "create ad-hoc account")
		}
	}
	user := &User{
		Email:        req.Email,
		Registration: &registration.Resource{URI: req.AccountURL},
		PrivateKey:   req.AccountKey,
	}
	cfg := lego.NewConfig(user)
	cfg.CADirURL = c.DirectoryURL
	cfg.Certificate.KeyType = keyType
	legoClient, err := lego.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create lego client")
	}
	err = legoClient.Challenge.SetHTTP01Provider(c.HTTP01Solver)
	if err != nil {
		return nil, errors.Wrap(err, "set challenge provider")
	}
	obtReq := certificate.ObtainRequest{
		Domains: req.Domains,
		Bundle:  req.Bundle,
	}
	certs, err := legoClient.Certificate.Obtain(obtReq)
	if err != nil {
		return nil, errors.Wrapf(err, "obtain certificates %s", req.Domains[0])
	}
	return &CertificateInfo{
		URL:               certs.CertURL,
		AccountURL:        user.Registration.URI,
		Certificate:       certs.Certificate,
		IssuerCertificate: certs.IssuerCertificate,
		PrivateKey:        certs.PrivateKey,
	}, nil
}

// CertificateRequest represents a request by an ACME protocol User to obtain
// or renew a certificate.
type CertificateRequest struct {
	Email      string            // Email address of the person responsible for the domains.
	AccountURL string            // URL of an already existing account; empty if no account exists.
	AccountKey crypto.PrivateKey // Private key of the account; don't confuse with the private key of a certificate.

	KeyType KeyType  // Type of key to use when requesting a certificate. Defaults to DefaultKeyType if not set.
	Domains []string // Domains for which a certificate is requested.
	Bundle  bool     // Bundle issuer certificate with issued certificate.
}

// CertificateInfo represents an ACME certificate along with its meta
// information.
type CertificateInfo struct {
	URL               string // URL of the certificate.
	AccountURL        string // URL of the certificate owner's account.
	Certificate       []byte // The actual certificate.
	PrivateKey        []byte // Private key used to generate the certificate.
	IssuerCertificate []byte // Certificate of the issuer of the certificate.
}
