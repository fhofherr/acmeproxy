package acme

import (
	"crypto"

	"github.com/fhofherr/acmeproxy/pkg/acme/internal/acme"
	"github.com/fhofherr/acmeproxy/pkg/acme/internal/challenge"
	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
	"github.com/pkg/errors"
)

// Client is an ACME protocol client capable of obtaining and renewing
// certificates.
type Client struct {
	DirectoryURL string
	HTTP01Solver *challenge.HTTP01Solver
}

// ObtainCertificate obtains a new certificate from the remote ACME server.
func (c *Client) ObtainCertificate(req CertificateRequest) (*CertificateInfo, error) {
	// TODO err if len(req.Domains) < 1
	u := req.newACMEUser()
	// TODO make keyType configurable per request
	legoClient, err := c.newLegoClient(u, certcrypto.RSA2048)
	if err != nil {
		return nil, errors.Wrap(err, "new lego client creation")
	}
	if err := u.Register(legoClient); err != nil {
		return nil, errors.Wrap(err, "user registration")
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
		AccountURL:        u.Registration.URI,
		Certificate:       certs.Certificate,
		IssuerCertificate: certs.IssuerCertificate,
		PrivateKey:        certs.PrivateKey,
	}, nil
}

func (c *Client) newLegoClient(u *acme.User, kt certcrypto.KeyType) (*lego.Client, error) {
	cfg := lego.NewConfig(u)
	cfg.CADirURL = c.DirectoryURL
	cfg.Certificate.KeyType = kt
	lc, err := lego.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create lego client")
	}
	err = lc.Challenge.SetHTTP01Provider(c.HTTP01Solver)
	if err != nil {
		return nil, errors.Wrap(err, "set challenge provider")
	}
	return lc, nil
}

// CertificateRequest represents a request by an ACME protocol User to obtain
// or renew a certificate.
type CertificateRequest struct {
	Email      string            // Email address of the person responsible for the domains.
	AccountURL string            // URL of an already existing account; empty if no account exists.
	PrivateKey crypto.PrivateKey // Private key for the certificate signing request.
	Domains    []string          // Domains for which a certificate is requested.
	Bundle     bool              // Bundle issuer certificate with issued certificate.
}

func (r CertificateRequest) newACMEUser() *acme.User {
	u := &acme.User{
		Email:      r.Email,
		PrivateKey: r.PrivateKey,
	}
	if r.AccountURL != "" {
		u.Registration = &registration.Resource{URI: r.AccountURL}
	}
	return u
}

// CertificateInfo represents an ACME certificate along with its meta
// information.
type CertificateInfo struct {
	URL               string
	AccountURL        string
	Certificate       []byte
	PrivateKey        []byte
	IssuerCertificate []byte
}
