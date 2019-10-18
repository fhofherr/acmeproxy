package acme

import (
	"crypto"
	"fmt"
	"io"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/go-acme/lego/lego"
	"github.com/google/uuid"
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
// of its users.
//
// Users use their unique user ID to register a domain. The Agent takes
// creates an ACME account for the user if necessary, and obtains a
// certificate for the domain. Later the user can obtain the certificate from
// the Agent using the domain name. Additionally it has to pass its unique
// user ID as a proof of ownership. This merely protects from obtaining
// certificates  for the wrong domain from within acmeproxy. The Agent does not
// implement any further mechanisms to ensure the user is actually allowed to
// retrieve certificates belonging to a domain.
type Agent struct {
	Domains      DomainRepository
	Users        UserRepository
	Certificates CertificateObtainer
	ACMEAccounts AccountCreator
}

// RegisterUser registers a new user of acme.Agent.
//
// The Agent creates an account for the user if it does not exist yet.
// The Agent uses the provided email as contact address for the account. If
// email is empty the Agent attempts to register an account without contact
// address.
//
// RegisterUser does nothing if the user has already been registered with
// the Agent.
func (a *Agent) RegisterUser(userID uuid.UUID, email string) error {
	const op errors.Op = "acme/agent.RegisterUser"

	_, err := a.Users.UpdateUser(userID, func(c *User) error {
		if !c.IsZero() {
			return nil
		}
		key, err := certutil.NewPrivateKey(certutil.EC256)
		if err != nil {
			return errors.New(op, fmt.Sprintf("new private key for user: %v", userID), err)
		}

		url, err := a.ACMEAccounts.CreateAccount(key, email)
		if err != nil {
			return errors.New(op, fmt.Sprintf("register account for user: %v", userID), err)
		}
		c.ID = userID
		c.Key = key
		c.AccountURL = url

		return nil
	})
	return errors.Wrap(err, op)
}

// RegisterDomain registers a new domain for the user uniquely identified by
// userID.
//
// Upon registration the Agent immediately attempts to obtain a certificate for
// the domain.
//
// RegisterDomain will silently ignore the domain if it was already registered
// with the same userID. It will return an error if the domain is already
// registered with a different userID.
func (a *Agent) RegisterDomain(userID uuid.UUID, domainName string) error {
	const op errors.Op = "acme/agent.RegisterDomain"

	user, err := a.Users.GetUser(userID)
	if err != nil {
		return errors.New(op, fmt.Sprintf("get user: %v", userID), err)
	}
	_, err = a.Domains.UpdateDomain(domainName, func(d *Domain) error {
		if !d.IsZero() {
			if userID != d.UserID {
				return errors.New(op, fmt.Sprintf("already registered: %s", domainName))
			}
			return nil
		}

		req := CertificateRequest{
			AccountURL: user.AccountURL,
			AccountKey: user.Key,
			KeyType:    DefaultKeyType,
			Domains:    []string{domainName},
			Bundle:     true,
		}
		ci, err := a.Certificates.ObtainCertificate(req)
		if err != nil {
			return errors.New(op, fmt.Sprintf("obtain certificate for domain: %s", domainName), err)
		}
		d.Name = domainName
		d.UserID = userID
		d.Certificate = ci.Certificate
		d.PrivateKey = ci.PrivateKey
		return nil
	})
	return errors.Wrap(err, op, fmt.Sprintf("update domain: %s", domainName))
}

// WriteCertificate writes the PEM encoded certificate for the domain to w.
//
// WriteCertificate returns an error if the domain was not registered, or was
// registered to a different userID. If the Agent has not yet obtained a
// certificate WriteCertificate returns an instance of RetryLater as error.
func (a *Agent) WriteCertificate(userID uuid.UUID, domainName string, w io.Writer) error {
	const op errors.Op = "acme/agent.WriteCertificate"

	domain, err := a.Domains.GetDomain(domainName)
	if err != nil {
		return errors.New(op, fmt.Sprintf("get domain: %s", domainName), err)
	}
	_, err = w.Write(domain.Certificate)
	return errors.Wrap(err, op, fmt.Sprintf("write certificate for domain: %s", domainName))
}

// WritePrivateKey writes the PEM encoded private key for the domain to w.
//
// WritePrivateKey returns an error if the domain was not registered, or was
// registered to a different userID. If the Agent has not yet obtained a
// certificate WritePrivateKey returns an instance of RetryLater as error.
func (a *Agent) WritePrivateKey(userID uuid.UUID, domainName string, w io.Writer) error {
	const op errors.Op = "acme/agent.WritePrivateKey"

	domain, err := a.Domains.GetDomain(domainName)
	if err != nil {
		return errors.New(op, fmt.Sprintf("get domain: %s", domainName), err)
	}
	_, err = w.Write(domain.PrivateKey)
	return errors.Wrap(err, op, fmt.Sprintf("write private key for domain: %s", domainName))
}
