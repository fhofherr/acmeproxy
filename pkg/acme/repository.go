package acme

import (
	"crypto"
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
	UserID      uuid.UUID
	Name        string
	Certificate []byte
	PrivateKey  []byte
}

// IsZero returns true if this domain is equal to its zero value.
func (d Domain) IsZero() bool {
	return reflect.DeepEqual(d, Domain{})
}

// UserRepository persists and retrieves information about the users of
// the Agent.
//
// The UpdateUser method atomically saves or updates a User. GetUser finds an
// user by its unique ID.
type UserRepository interface {
	UpdateUser(uuid.UUID, func(c *User) error) (User, error)
	GetUser(uuid.UUID) (User, error)
}

// User represents a user of the Agent.
type User struct {
	ID         uuid.UUID         // Unique identifier of the user.
	Key        crypto.PrivateKey // Private key used to identify the account with the ACME certificate authority.
	AccountURL string            // URL of the user's account at the ACME certificate authority.
}

// IsZero returns true if this user is equal to its zero value.
func (c User) IsZero() bool {
	return reflect.DeepEqual(c, User{})
}
