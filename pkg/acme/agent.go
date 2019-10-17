package acme

import (
	"crypto"
	"io"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/go-acme/lego/lego"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// DefaultDirectoryURL points to Let's Encrypt's production directory.
const DefaultDirectoryURL = lego.LEDirectoryProduction

// DefaultKeyType is the default key type to use if the CertificateRequest does
// not specify one.
const DefaultKeyType = certutil.RSA2048

// CertificateRequest represents a request by an ACME protocol User to obtain
// or renew a certificate.
type CertificateRequest struct {
	Email      string            // Email address of the person responsible for the domains.
	AccountURL string            // URL of an already existing account; empty if no account exists.
	AccountKey crypto.PrivateKey // Private key of the account; don't confuse with the private key of a certificate.

	KeyType certutil.KeyType // Type of key to use when requesting a certificate. Defaults to DefaultKeyType if not set.
	Domains []string         // Domains for which a certificate is requested.
	Bundle  bool             // Bundle issuer certificate with issued certificate.
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

// CertificateObtainer wraps the ObtainCertificate method which obtains
// a certificate for a specific domain from an ACME certificate authority.
type CertificateObtainer interface {
	ObtainCertificate(CertificateRequest) (*CertificateInfo, error)
}

// AccountCreator wraps the CreateAccount method which creates an new
// account at the ACME certificate authority.
type AccountCreator interface {
	CreateAccount(key crypto.PrivateKey, email string) (string, error)
}

// Agent takes care of obtaining and renewing ACME certificates for the domains
// of its clients.
//
// Clients use their unique client ID to register a domain. The Agent takes
// creates an ACME account for the client if necessary, and obtains a
// certificate for the domain. Later the client can obtain the certificate from
// the Agent using the domain name. Additionally it has to pass its unique
// client ID as a proof of ownership. This merely protects from obtaining
// certificates  for the wrong domain from within acmeproxy. The Agent does not
// implement any further mechanisms to ensure the client is actually allowed to
// retrieve certificates belonging to a domain.
type Agent struct {
	Domains      DomainRepository
	Clients      ClientRepository
	Certificates CertificateObtainer
	ACMEAccounts AccountCreator
}

// RegisterClient registers a new client of acme.Agent.
//
// The Agent creates an account for the client if it does not exist yet.
// The Agent uses the provided email as contact address for the account. If
// email is empty the Agent attempts to register an account without contact
// address.
//
// RegisterClient does nothing if the client has already been registered with
// the Agent.
func (a *Agent) RegisterClient(clientID uuid.UUID, email string) error {
	_, err := a.Clients.UpdateClient(clientID, func(c *Client) error {
		if !c.IsZero() {
			return nil
		}
		key, err := certutil.NewPrivateKey(certutil.EC256)
		if err != nil {
			return errors.Wrapf(err, "new private key for client: %v", clientID)
		}

		url, err := a.ACMEAccounts.CreateAccount(key, email)
		if err != nil {
			return errors.Wrapf(err, "register account for client: %v", clientID)
		}
		c.ID = clientID
		c.Key = key
		c.AccountURL = url

		return nil
	})
	return errors.Wrapf(err, "register client")
}

// RegisterDomain registers a new domain for the client uniquely identified by
// clientID.
//
// Upon registration the Agent immediately attempts to obtain a certificate for
// the domain.
//
// RegisterDomain will silently ignore the domain if it was already registered
// with the same clientID. It will return an error if the domain is already
// registered with a different clientID.
func (a *Agent) RegisterDomain(clientID uuid.UUID, domainName string) error {
	client, err := a.Clients.GetClient(clientID)
	if err != nil {
		return errors.Wrapf(err, "get client: %v", clientID)
	}
	_, err = a.Domains.UpdateDomain(domainName, func(d *Domain) error {
		if !d.IsZero() {
			if clientID != d.ClientID {
				return errors.Errorf("domain already registered: %s", domainName)
			}
			return nil
		}

		req := CertificateRequest{
			AccountURL: client.AccountURL,
			AccountKey: client.Key,
			KeyType:    DefaultKeyType,
			Domains:    []string{domainName},
			Bundle:     true,
		}
		ci, err := a.Certificates.ObtainCertificate(req)
		if err != nil {
			return errors.Wrapf(err, "obtain certificate for domain: %s", domainName)
		}
		d.Name = domainName
		d.ClientID = clientID
		d.Certificate = ci.Certificate
		d.PrivateKey = ci.PrivateKey
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "update domain: %s", domainName)
	}
	return nil
}

// WriteCertificate writes the PEM encoded certificate for the domain to w.
//
// WriteCertificate returns an error if the domain was not registered, or was
// registered to a different clientID. If the Agent has not yet obtained a
// certificate WriteCertificate returns an instance of RetryLater as error.
func (a *Agent) WriteCertificate(clientID uuid.UUID, domainName string, w io.Writer) error {
	domain, err := a.Domains.GetDomain(domainName)
	if err != nil {
		return errors.Wrapf(err, "get domain: %s", domainName)
	}
	_, err = w.Write(domain.Certificate)
	return errors.Wrapf(err, "write certificate for domain: %s", domainName)
}

// WritePrivateKey writes the PEM encoded private key for the domain to w.
//
// WritePrivateKey returns an error if the domain was not registered, or was
// registered to a different clientID. If the Agent has not yet obtained a
// certificate WritePrivateKey returns an instance of RetryLater as error.
func (a *Agent) WritePrivateKey(clientID uuid.UUID, domainName string, w io.Writer) error {
	domain, err := a.Domains.GetDomain(domainName)
	if err != nil {
		return errors.Wrapf(err, "get domain: %s", domainName)
	}
	_, err = w.Write(domain.PrivateKey)
	return errors.Wrapf(err, "write private key for domain: %s", domainName)
}

// ErrRetryLater signals that this error may vanish with time and that a caller
// should retry the operation later.
type ErrRetryLater struct{}

func (err ErrRetryLater) Error() string {
	return ""
}

// WaitDuration returns the duration after which the operation may be retried.
func (err ErrRetryLater) WaitDuration() time.Duration {
	return 0
}
