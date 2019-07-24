package acme

import (
	"crypto"

	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
	"github.com/pkg/errors"
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

// Register creates a new reqistration with the ACME certificate authority
// and sets u.Registration. Does nothing if u.Registration is already set
// to some value.
func (u *User) Register(lc *lego.Client) error {
	var err error

	if u.Registration != nil {
		return nil
	}
	opts := registration.RegisterOptions{TermsOfServiceAgreed: true}
	u.Registration, err = lc.Registration.Register(opts)
	if err != nil {
		return errors.Wrapf(err, "register user: %s", u.Email)
	}
	return nil
}
