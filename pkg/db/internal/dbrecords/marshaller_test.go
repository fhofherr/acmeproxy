package dbrecords_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/db/internal/dbrecords"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMarshalAndUnmarshalUUID(t *testing.T) {
	id := uuid.Must(uuid.NewRandom())
	bs, err := dbrecords.MarshalBinary(id)
	assert.NoError(t, err)
	actual := uuid.UUID{}
	err = dbrecords.UnmarshalBinary(bs, &actual)
	assert.NoError(t, err)
	assert.Equal(t, id, actual)
}

func TestMarshalAndUnmarshalString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		convert  func(string) interface{}
	}{
		{
			name:     "string",
			expected: "some string",
			convert: func(s string) interface{} {
				return s
			},
		},
		{
			name:     "*string",
			expected: "some string pointer",
			convert: func(s string) interface{} {
				return &s
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var actual string

			bs, err := dbrecords.MarshalBinary(tt.convert(tt.expected))
			if !assert.NoError(t, err) {
				return
			}
			err = dbrecords.UnmarshalBinary(bs, &actual)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestMarshalAndUnmarshalClients(t *testing.T) {
	tests := []struct {
		name     string
		keyType  certutil.KeyType
		keyFile  string
		toSource func(*acme.Client) interface{}
	}{
		{
			name:    "acme.Client",
			keyType: certutil.EC256,
			keyFile: "acme_client.pem",
			toSource: func(c *acme.Client) interface{} {
				return *c
			},
		},
		{
			name:    "pointer to acme.Client",
			keyType: certutil.EC384,
			keyFile: "pointer_to_acme_client.pem",
			// This is the default behavior. For clarity reasons this function
			// re-states it.
			toSource: func(c *acme.Client) interface{} {
				return c
			},
		},
		{
			name:    "client with ECDSA Key",
			keyType: certutil.EC256,
			keyFile: "client_with_ecdsa_key.pem",
		},
		{
			name:    "client with RSA Key",
			keyType: certutil.RSA2048,
			keyFile: "client_with_rsa_key.pem",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var source interface{}

			keyFile := filepath.Join("testdata", t.Name(), tt.keyFile)
			if *testsupport.FlagUpdate {
				certutil.WritePrivateKeyForTesting(t, keyFile, tt.keyType, true)
			}

			client := &acme.Client{
				ID:         uuid.Must(uuid.NewRandom()),
				Key:        certutil.KeyMust(certutil.ReadPrivateKeyFromFile(tt.keyType, keyFile, true)),
				AccountURL: "https://example.com/some/account",
			}
			source = client
			if tt.toSource != nil {
				source = tt.toSource(client)
			}
			bs, err := dbrecords.MarshalBinary(source)
			if !assert.NoError(t, err) {
				return
			}
			target := &acme.Client{}
			err = dbrecords.UnmarshalBinary(bs, target)
			if !assert.NoError(t, err) {
				return
			}
			assertDomainObjectsEqual(t, source, target)
		})
	}
}

func TestMarshalAndUnmarshalDomain(t *testing.T) {
	tests := []struct {
		name     string
		toSource func(*acme.Domain) interface{}
	}{
		{
			name: "acme.Domain",
			toSource: func(domain *acme.Domain) interface{} {
				return *domain
			},
		},
		{
			name: "pointer to acme.Domain",
			toSource: func(domain *acme.Domain) interface{} {
				return domain
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			domainName := "example.com"
			keyFile := filepath.Join("testdata", t.Name(), "key.pem")
			certFile := filepath.Join("testdata", t.Name(), "certificate.pem")
			if *testsupport.FlagUpdate {
				pk := certutil.WritePrivateKeyForTesting(t, keyFile, certutil.RSA2048, true)
				certutil.WriteCertificateForTesting(t, certFile, domainName, pk, true)
			}
			keyBytes, err := ioutil.ReadFile(keyFile)
			if err != nil {
				t.Fatal(err)
			}
			certBytes, err := ioutil.ReadFile(certFile)
			if err != nil {
				t.Fatal(err)
			}
			domain := &acme.Domain{
				ClientID:    uuid.Must(uuid.NewRandom()),
				Name:        domainName,
				Certificate: certBytes,
				PrivateKey:  keyBytes,
			}
			var source interface{} = domain
			if tt.toSource != nil {
				source = tt.toSource(domain)
			}
			bs, err := dbrecords.MarshalBinary(source)
			assert.NoError(t, err)
			target := &acme.Domain{}
			err = dbrecords.UnmarshalBinary(bs, target)
			assert.NoError(t, err)
			assertDomainObjectsEqual(t, domain, target)
		})
	}
}

func assertDomainObjectsEqual(t *testing.T, expected, actual interface{}) {
	switch v := expected.(type) {
	case acme.Client:
		assert.Equal(t, &v, actual)
	default:
		assert.Equal(t, expected, actual)
	}
}
