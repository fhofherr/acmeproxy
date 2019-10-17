package acmeclient

import (
	"crypto"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
	"github.com/pkg/errors"
)

// Client is an ACME protocol client capable of obtaining and renewing
// certificates.
type Client struct {
	DirectoryURL string
	HTTP01Solver HTTP01Solver
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
func (c *Client) ObtainCertificate(req acme.CertificateRequest) (*acme.CertificateInfo, error) {
	// TODO err if len(req.Domains) < 1
	keyType, err := legoKeyType(req.KeyType)
	if err != nil {
		return nil, errors.Wrap(err, "determine lego key type")
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
	err = legoClient.Challenge.SetHTTP01Provider(&c.HTTP01Solver)
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
	return &acme.CertificateInfo{
		URL:               certs.CertURL,
		AccountURL:        user.Registration.URI,
		Certificate:       certs.Certificate,
		IssuerCertificate: certs.IssuerCertificate,
		PrivateKey:        certs.PrivateKey,
	}, nil
}
