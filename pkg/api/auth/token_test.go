package auth_test

import (
	"crypto"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/fhofherr/acmeproxy/pkg/api/auth"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestNewToken(t *testing.T) {
	tests := []struct {
		name    string
		alg     auth.Algorithm
		keyType certutil.KeyType
		err     error
	}{
		{
			name: "returns error for unknown algorithm",
			alg:  auth.Algorithm(-1),
			err:  errors.New("unknown algorithm"),
		},
		{
			name:    "returns error for unsuitable key",
			alg:     auth.ES256,
			keyType: certutil.RSA2048,
			err:     errors.New("sign token", fmt.Errorf("key is of invalid type")),
		},
		{
			name:    "returns signed token",
			alg:     auth.ES256,
			keyType: certutil.EC256,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			keyFile := filepath.Join("testdata", t.Name(), "key.pem")
			if *testsupport.FlagUpdate {
				certutil.WritePrivateKeyForTesting(t, keyFile, tt.keyType, true)
			}
			key, err := certutil.ReadPrivateKeyFromFile(tt.keyType, keyFile, true)
			if err != nil {
				t.Fatal(err)
			}
			cs := &auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "jdoe",
				},
			}
			token, err := auth.NewToken(cs, tt.alg, key)
			assert.Truef(t, errors.Is(err, tt.err), "expected %v; got %v", tt.err, err)
			if tt.err != nil {
				return
			}
			assert.NotEmpty(t, token)
		})
	}
}

func TestParseToken_InvalidAlgorithm(t *testing.T) {
	_, err := auth.ParseToken("", auth.Algorithm(-1), nil)
	assert.True(t, errors.Is(err, errors.New(errors.Unauthorized, "unknown algorithm")))
}

func TestParseToken(t *testing.T) {
	tests := []struct {
		name        string
		tokenAlg    auth.Algorithm
		parseAlg    auth.Algorithm
		useParseKey bool
		err         error
		claims      *auth.Claims
	}{
		{
			name:     "returns error on signing algorithm mismatch",
			parseAlg: auth.ES512,
			err:      errors.New(errors.Unauthorized, "signing algorithm mismatch"),
		},
		{
			name:        "returns error on invalid signature",
			useParseKey: true,
			claims: &auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "jdoe",
				},
			},
			err: errors.New(errors.Unauthorized, "invalid signature"),
		},
		{
			name: "returns claims of parsed token",
			claims: &auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "jdoe",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tokenKeyFile := filepath.Join("testdata", t.Name(), "key.pem")
			parseKeyFile := filepath.Join("testdata", t.Name(), "parseKey.pem")
			tokenKeyType := keyTypeForAlg(t, tt.tokenAlg)
			parseKeyType := keyTypeForAlg(t, tt.parseAlg)
			tokenFile := filepath.Join("testdata", t.Name(), "token")

			if *testsupport.FlagUpdate {
				tokenKey := certutil.WritePrivateKeyForTesting(t, tokenKeyFile, tokenKeyType, true)
				token, err := auth.NewToken(tt.claims, tt.tokenAlg, tokenKey)
				if err != nil {
					t.Fatal(err)
				}
				if err := ioutil.WriteFile(tokenFile, []byte(token), 0644); err != nil {
					t.Fatal(err)
				}
				if tt.useParseKey {
					certutil.WritePrivateKeyForTesting(t, parseKeyFile, parseKeyType, true)
				}
			}

			tokenKey, err := certutil.ReadPrivateKeyFromFile(tokenKeyType, tokenKeyFile, true)
			if err != nil {
				t.Fatal(err)
			}
			token, err := ioutil.ReadFile(tokenFile)
			if err != nil {
				t.Fatal(err)
			}
			pk := publicKey(t, tokenKey)
			if tt.useParseKey {
				parseKey, err := certutil.ReadPrivateKeyFromFile(parseKeyType, parseKeyFile, true)
				if err != nil {
					t.Fatal(err)
				}
				pk = publicKey(t, parseKey)
			}
			claims, err := auth.ParseToken(string(token), tt.parseAlg, pk)

			assert.Truef(t, errors.Is(err, tt.err), "expected %v; got %v", tt.err, err)
			if tt.err != nil {
				return
			}
			assert.Equal(t, tt.claims, claims)
		})
	}
}

func keyTypeForAlg(t *testing.T, alg auth.Algorithm) certutil.KeyType {
	switch alg {
	case auth.ES256:
		return certutil.EC256
	case auth.ES512:
		return certutil.EC521
	default:
		t.Fatalf("Can't determine key type for algorithm")
		return certutil.KeyType(-1)
	}
}

func publicKey(t *testing.T, key crypto.PrivateKey) crypto.PublicKey {
	s, ok := key.(crypto.Signer)
	if !ok {
		t.Fatal("key is not a crypto.Signer")
	}
	return s.Public()
}

func TestCheckRoles(t *testing.T) {
	tests := []struct {
		name          string
		tokenRoles    []auth.Role
		requiredRoles []auth.Role
		err           error
	}{
		{
			name:          "no roles in token",
			requiredRoles: []auth.Role{auth.Admin},
			err:           errors.New(errors.Unauthorized, "no roles present"),
		},
		{
			name:       "no required roles passed",
			tokenRoles: []auth.Role{auth.Admin},
			err:        errors.New(errors.InvalidArgument),
		},
		{
			name:          "none of the token roles matches",
			tokenRoles:    []auth.Role{auth.Role("some role")},
			requiredRoles: []auth.Role{auth.Role("another role")},
			err:           errors.New(errors.Unauthorized, "at least one role required: [another role]"),
		},
		{
			name:          "one of the token roles matches",
			tokenRoles:    []auth.Role{auth.Role("some role"), auth.Admin},
			requiredRoles: []auth.Role{auth.Admin},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			claims := &auth.Claims{
				Roles: tt.tokenRoles,
			}
			err := claims.CheckRoles(tt.requiredRoles...)
			errors.AssertMatches(t, tt.err, err)
		})
	}
}

func TestCheckRoles_NilCaller(t *testing.T) {
	var claims *auth.Claims
	template := errors.New(errors.Unauthorized, "no token present")
	err := claims.CheckRoles(auth.Admin)
	errors.AssertMatches(t, template, err)
}
