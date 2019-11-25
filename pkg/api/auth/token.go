package auth

import (
	"crypto"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/fhofherr/acmeproxy/pkg/errors"
)

// Algorithm represents the singing algorithm used to sign the tokens issued
// by the auth package. Only asymmetrically singed tokens are supported.
type Algorithm int

const (
	// ES256 represents the ES256 signing algorithm
	ES256 Algorithm = iota
	// ES512 represents the ES512 signing algorithm
	ES512
)

func (a Algorithm) signingMethod() (jwt.SigningMethod, error) {
	const op errors.Op = "auth/algorithm.SigningMethod"

	switch a {
	case ES256:
		return jwt.SigningMethodES256, nil
	case ES512:
		return jwt.SigningMethodES256, nil
	default:
		return nil, errors.New(op, "unknown algorithm")
	}
}

func (a Algorithm) methodOk(method jwt.SigningMethod) bool {
	switch a {
	case ES256:
		return method == jwt.SigningMethodES256
	case ES512:
		return method == jwt.SigningMethodES512
	default:
		return false
	}
}

// Role is the type of the role names in the JWT.
type Role string

const (
	// Admin signals that the bearer of the token is an administrator
	Admin Role = "admin"
)

// Claims represents the claims of a JWT token a valid client of acmeproxy
// may present. In contains the JWT standard claims as well as additional
// claims which might be necessary in the future.
type Claims struct {
	jwt.StandardClaims
	Roles []Role `json:"roles,omitempty"`
}

// NewToken creates and signs the a JWT containing the passed claims.
//
// The resulting token signed with key. NewToken returns an error if the key is
// not suitable for the passed signing algorithm.
func NewToken(c *Claims, alg Algorithm, key crypto.PrivateKey) (string, error) {
	const op errors.Op = "auth/NewToken"

	sm, err := alg.signingMethod()
	if err != nil {
		return "", errors.New(op, err)
	}
	token := jwt.NewWithClaims(sm, c)
	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", errors.New(op, "sign token", err)
	}
	return tokenStr, nil
}

// ParseToken parses the passed token and verifies its signature. It ensures
// that only tokens whose signing algorithm matches the passed algorithm are
// accepted.
func ParseToken(token string, alg Algorithm, key crypto.PublicKey) (*Claims, error) {
	const op errors.Op = "auth/ParseToken"
	var claims Claims

	// Validate the algorithm and check if it belongs to a supported singing
	// method
	if _, err := alg.signingMethod(); err != nil {
		// Don't wrap err here as we do not want deeply nested errors.
		// Just responding with an adequate message seems more elegant.
		return &claims, errors.New(op, errors.Unauthorized, "unknown algorithm")
	}
	_, err := jwt.ParseWithClaims(token, &claims, func(tok *jwt.Token) (interface{}, error) {
		if !alg.methodOk(tok.Method) {
			return nil, errors.New(op, errors.Unauthorized, "signing algorithm mismatch")
		}
		return key, nil
	})
	if err := handleValidationError(op, err); err != nil {
		return &claims, errors.New(op, err)
	}
	return &claims, nil
}

func handleValidationError(op errors.Op, err error) error {
	if err == nil {
		return nil
	}

	var vErr *jwt.ValidationError

	if !errors.As(err, &vErr) {
		return errors.New(op, err)
	}
	if (vErr.Errors & jwt.ValidationErrorSignatureInvalid) > 0 {
		return errors.New(op, errors.Unauthorized, "invalid signature")
	}
	// Unwrap the validation error if vErr.Inner is an instance of errors/Error.
	// The way errors/Error.Is is defined ensures we can test against an empty
	// instance.
	if vErr.Inner != nil && errors.Is(vErr.Inner, &errors.Error{}) {
		return vErr.Inner
	}
	return errors.New(op, err)
}

// CheckRoles checks if the passed claims object contains at least of the
// expected roles. It returns an error of kind Unauthorized if none of the
// expected roles could be found.
func (c *Claims) CheckRoles(roles ...Role) error {
	const op errors.Op = "auth/claims.CheckRoles"

	if len(roles) == 0 {
		return errors.New(op, errors.InvalidArgument, "no roles to check")
	}
	if c == nil {
		return errors.New(op, errors.Unauthorized, "no token present")
	}
	if len(c.Roles) == 0 {
		return errors.New(op, errors.Unauthorized, "no roles present")
	}
	rolesSet := make(map[Role]struct{}, len(roles))
	for _, r := range roles {
		rolesSet[r] = struct{}{}
	}
	foundRole := false
	for _, r := range c.Roles {
		if _, ok := rolesSet[r]; ok {
			foundRole = true
		}
	}
	if !foundRole {
		msg := fmt.Sprintf("at least one role required: %v", roles)
		return errors.New(op, errors.Unauthorized, msg)
	}
	return nil
}
