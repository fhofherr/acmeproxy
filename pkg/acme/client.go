package acme

import (
	"crypto"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
	"github.com/pkg/errors"
)

// Client is an ACME protocol client capable of obtaining and renewing
// certificates.
type Client struct {
	DirectoryURL     string
	ChallengeHandler *HTTP01Handler
}

// ObtainCertificate obtains a new certificate from the remote ACME server.
func (c *Client) ObtainCertificate(req CertificateRequest) (*CertificateInfo, error) {
	// TODO err if len(req.Domains) < 1
	// TODO as per the lego documentation the key is optional, how can we make
	//      this happen? Currently it does not work.
	u := &user{
		Email:      req.Email,
		PrivateKey: req.Key,
	}
	cfg := lego.NewConfig(u)
	cfg.CADirURL = c.DirectoryURL
	// TODO make this configurable per request
	cfg.Certificate.KeyType = certcrypto.RSA2048
	legoClient, err := lego.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create lego client")
	}
	err = legoClient.Challenge.SetHTTP01Provider(c.ChallengeHandler)
	if err != nil {
		return nil, errors.Wrap(err, "set challenge provider")
	}
	// TODO test what happens if we don't create an account. Do we need to save
	//      the registration.
	if req.CreateAccount {
		opts := registration.RegisterOptions{TermsOfServiceAgreed: true}
		reg, err := legoClient.Registration.Register(opts)
		if err != nil {
			return nil, errors.Wrapf(err, "register user %s", req.Email)
		}
		u.Registration = reg
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
		Certificate:       certs.Certificate,
		IssuerCertificate: certs.IssuerCertificate,
		PrivateKey:        certs.PrivateKey,
	}, nil
}

// CertificateRequest represents a request by an ACME protocol user to obtain
// or renew a certificate.
type CertificateRequest struct {
	Email         string            // Email address of the person responsible for the domains.
	Key           crypto.PrivateKey // Private key for the certificate signing request.
	Domains       []string          // Domains for which a certificate is requested.
	CreateAccount bool              // Set to true if Email is used for the first time and requires a new account.
	Bundle        bool              // Bundle issuer certificate with issued certificate.
}

// CertificateInfo represents an ACME certificate along with its meta
// information.
type CertificateInfo struct {
	URL               string
	Certificate       []byte
	PrivateKey        []byte
	IssuerCertificate []byte
}

type user struct {
	Email        string
	Registration *registration.Resource
	PrivateKey   crypto.PrivateKey
}

// GetEmail returns the users email.
func (u *user) GetEmail() string {
	return u.Email
}

// GetRegistration returns the users registration.
func (u *user) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the users private key.
func (u *user) GetPrivateKey() crypto.PrivateKey {
	return u.PrivateKey
}
