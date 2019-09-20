package acmeclient_test

import (
	"strings"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme/acmeclient"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)
	tests := []struct {
		name  string
		email string
	}{
		{name: "create account without email"},
		{name: "create account with email", email: "jane.doe@example.com"},
	}

	fx, tearDown := acmeclient.NewTestFixture(t, testsupport.ChallengeServerPort)
	defer tearDown()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// TODO(fhofherr) read from golden file instead
			accountKey := certutil.KeyMust(certutil.NewPrivateKey(certutil.EC256))
			accountURL, err := fx.Client.CreateAccount(accountKey, tt.email)
			assert.NoError(t, err)
			assert.NotEmpty(t, accountURL)
			assert.Truef(
				t,
				strings.HasPrefix(accountURL, fx.Pebble.AccountURLPrefix()),
				"accountURL %s did not start with %s",
				accountURL,
				fx.Pebble.AccountURLPrefix())
		})
	}
}

func TestObtainCertificate(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)

	fx, tearDown := acmeclient.NewTestFixture(t, testsupport.ChallengeServerPort)
	defer tearDown()

	tests := []struct {
		acmeclient.CertificateRequest
		name string
	}{
		{
			name: "obtain certificate without account",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA2048@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: certutil.KeyMust(certutil.NewPrivateKey(certutil.EC256)),
			},
		},
		{
			name: "obtain RSA4096 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA4096@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: certutil.KeyMust(certutil.NewPrivateKey(certutil.EC256)),
				KeyType:    certutil.RSA4096,
			},
		},
		{
			name: "obtain RSA8192 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+RSA8192@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: certutil.KeyMust(certutil.NewPrivateKey(certutil.EC256)),
				KeyType:    certutil.RSA8192,
			},
		},
		{
			name: "obtain EC256 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+EC256@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: certutil.KeyMust(certutil.NewPrivateKey(certutil.EC256)),
				KeyType:    certutil.EC256,
			},
		},
		{
			name: "obtain EC384 certificate",
			CertificateRequest: acmeclient.CertificateRequest{
				Email:      "john.doe+EC384@example.com",
				Domains:    []string{"www.example.com"},
				Bundle:     true,
				AccountKey: certutil.KeyMust(certutil.NewPrivateKey(certutil.EC256)),
				KeyType:    certutil.EC384,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			certInfo, err := fx.Client.ObtainCertificate(tt.CertificateRequest)
			if !assert.NoError(t, err) {
				return
			}
			assert.NotEmpty(t, certInfo.URL)
			assert.NotEmpty(t, certInfo.AccountURL)
			for _, domain := range tt.CertificateRequest.Domains {
				certutil.AssertCertificateValid(t, domain, certInfo.IssuerCertificate, certInfo.Certificate)
				certutil.AssertKeyBelongsToCertificate(t, tt.KeyType, certInfo.Certificate, certInfo.PrivateKey)
				fx.Pebble.AssertIssuedByPebble(t, domain, certInfo.Certificate)
			}
		})
	}
}

func TestObtainCertificateWithPreExistingAccount(t *testing.T) {
	testsupport.SkipIfPebbleDisabled(t)

	fx, tearDown := acmeclient.NewTestFixture(t, testsupport.ChallengeServerPort)
	defer tearDown()

	domain := "www.example.com"
	accountKey := certutil.KeyMust(certutil.NewPrivateKey(certutil.EC256))
	accountURL, err := fx.Client.CreateAccount(accountKey, "jane.doe@example.com")
	assert.NoError(t, err)

	req := acmeclient.CertificateRequest{
		Email:      "jane.doe+RSA2048@example.com",
		AccountURL: accountURL,
		Domains:    []string{domain},
		Bundle:     true,
		AccountKey: accountKey,
		KeyType:    certutil.RSA2048,
	}
	ci, err := fx.Client.ObtainCertificate(req)
	assert.NoError(t, err)
	fx.Pebble.AssertIssuedByPebble(t, domain, ci.Certificate)
}
