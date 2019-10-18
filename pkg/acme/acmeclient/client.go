package acmeclient

import (
	"crypto"
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
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
	const op errors.Op = "acmeclient/client.CreateAccount"

	user := &User{
		Email:      email,
		PrivateKey: accountKey,
	}
	cfg := lego.NewConfig(user)
	cfg.CADirURL = c.DirectoryURL
	legoClient, err := lego.NewClient(cfg)
	if err != nil {
		return "", errors.New(op, "create client for new ACME account", err)
	}
	err = user.Register(legoClient)
	if err != nil {
		return "", errors.New(op, "register new ACME account", err)
	}
	return user.Registration.URI, nil
}

// ObtainCertificate obtains a new certificate from the remote ACME server.
func (c *Client) ObtainCertificate(req acme.CertificateRequest) (*acme.CertificateInfo, error) {
	const op errors.Op = "acmeclient/client.ObtainCertificate"

	if len(req.Domains) < 1 {
		return nil, errors.New(op, errors.InvalidArgument, "no domains")
	}
	keyType, err := legoKeyType(req.KeyType)
	if err != nil {
		return nil, errors.New(op, "determine lego key type", err)
	}
	if req.AccountURL == "" {
		var err error
		req.AccountURL, err = c.CreateAccount(req.AccountKey, req.Email)
		if err != nil {
			return nil, errors.New(op, "create ad-hoc account", err)
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
		return nil, errors.New(op, "create lego client", err)
	}
	err = legoClient.Challenge.SetHTTP01Provider(&c.HTTP01Solver)
	if err != nil {
		return nil, errors.New(op, "set challenge provider", err)
	}
	obtReq := certificate.ObtainRequest{
		Domains: req.Domains,
		Bundle:  req.Bundle,
	}
	certs, err := legoClient.Certificate.Obtain(obtReq)
	if err != nil {
		return nil, errors.New(op, fmt.Sprintf("obtain certificates %s", req.Domains[0]), err)
	}
	return &acme.CertificateInfo{
		URL:               certs.CertURL,
		AccountURL:        user.Registration.URI,
		Certificate:       certs.Certificate,
		IssuerCertificate: certs.IssuerCertificate,
		PrivateKey:        certs.PrivateKey,
	}, nil
}
