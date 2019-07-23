package acme

import (
	"crypto"

	"github.com/go-acme/lego/registration"
)

// User represents an user of the ACME certificate authority.
//
// It implements https://godoc.org/github.com/go-acme/lego/registration#User.
type User struct {
	Email        string
	Registration *registration.Resource
	PrivateKey   crypto.PrivateKey
}

// GetEmail returns the users email.
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the users registration.
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the users private key.
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.PrivateKey
}
