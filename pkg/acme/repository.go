package acme

import (
	"crypto"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

// DomainRepository persists and retrieves information about the domains managed
// by the Agent.
//
// The UpdateDomain method atomically saves or updates a Domain. GetDomain finds
// a domain by its domain name.
type DomainRepository interface {
	UpdateDomain(string, func(d *Domain) error) (Domain, error)
	GetDomain(string) (Domain, error)
}

// Domain represents a domain managed by the Agent.
type Domain struct {
	ClientID    uuid.UUID
	Name        string
	Certificate []byte
	PrivateKey  []byte
}

// IsZero returns true if this domain is equal to its zero value.
func (d Domain) IsZero() bool {
	return reflect.DeepEqual(d, Domain{})
}

// DomainNotFound is an error returned by domain repositories in order to signal
// that a domain was not found.
type DomainNotFound struct {
	DomainName string
}

func (d DomainNotFound) Error() string {
	return fmt.Sprintf("domain not found: %s", d.DomainName)
}

func isDomainNotFound(err error) bool {
	_, ok := err.(DomainNotFound)
	return ok
}

// ClientRepository persists and retrieves information about the clients of
// the Agent.
//
// The UpdateClient atomically saves or updates a Client. GetClient finds a
// client by its unique ID.
type ClientRepository interface {
	UpdateClient(uuid.UUID, func(c *Client) error) (Client, error)
	GetClient(uuid.UUID) (Client, error)
}

// Client represents a client of the Agent.
//
// TODO(fhofherr) consider renaming Client to User or something better. Client
//                clashes with the ACME client.
type Client struct {
	ID         uuid.UUID         // Unique identifier of the client.
	Key        crypto.PrivateKey // Private key used to identify the account with the ACME certificate authority.
	AccountURL string            // URL of the client's account at the ACME certificate authority.
}

// IsZero returns true if this client is equal to its zero value.
func (c Client) IsZero() bool {
	return reflect.DeepEqual(c, Client{})
}

// ClientNotFound is an error returned by client repositories in order to signal
// that a client was not found.
type ClientNotFound struct {
	ClientID uuid.UUID
}

func (c ClientNotFound) Error() string {
	return fmt.Sprintf("client not found: %v", c.ClientID)
}
