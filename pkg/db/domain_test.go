package db_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveNewDomain(t *testing.T) {
	domainName := "example.com"
	certFile := filepath.Join("testdata", t.Name(), "certificate.pem")
	keyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
	if *db.FlagUpdate {
		pk := certutil.WritePrivateKeyForTesting(t, keyFile, certutil.EC256, true)
		certutil.WriteCertificateForTesting(t, certFile, domainName, pk, true)
	}

	fx := db.NewTestFixture(t)
	defer fx.Close()
	domainRepository := fx.DB.DomainRepository()

	certBytes := readFile(t, certFile)
	keyBytes := readFile(t, keyFile)
	domain := acme.Domain{
		ClientID:    uuid.Must(uuid.NewRandom()),
		Name:        domainName,
		Certificate: certBytes,
		PrivateKey:  keyBytes,
	}
	actual, err := domainRepository.UpdateDomain(domainName, func(d *acme.Domain) error {
		d.ClientID = domain.ClientID
		d.Name = domain.Name
		d.Certificate = domain.Certificate
		d.PrivateKey = domain.PrivateKey
		return nil
	})
	assert.NoError(t, err)
	saved, err := domainRepository.GetDomain(domainName)
	assert.NoError(t, err)
	assert.Equal(t, domain, actual)
	assert.Equal(t, domain, saved)
}

func TestUpdateDomain(t *testing.T) {
	domainName := "example.com"
	initialCertFile := filepath.Join("testdata", t.Name(), "initial_certificate.pem")
	updatedCertFile := filepath.Join("testdata", t.Name(), "updated_certificate.pem")
	initialKeyFile := filepath.Join("testdata", t.Name(), "initial_private_key.pem")
	updatedKeyFile := filepath.Join("testdata", t.Name(), "updated_private_key.pem")
	if *db.FlagUpdate {
		pk := certutil.WritePrivateKeyForTesting(t, initialKeyFile, certutil.EC256, true)
		certutil.WriteCertificateForTesting(t, initialCertFile, domainName, pk, true)
		pk = certutil.WritePrivateKeyForTesting(t, updatedKeyFile, certutil.EC256, true)
		certutil.WriteCertificateForTesting(t, updatedCertFile, domainName, pk, true)
	}

	fx := db.NewTestFixture(t)
	defer fx.Close()
	domainRepository := fx.DB.DomainRepository()

	newDomain, err := domainRepository.UpdateDomain(domainName, func(d *acme.Domain) error {
		d.Name = domainName
		d.ClientID = uuid.Must(uuid.NewRandom())
		d.Certificate = readFile(t, initialCertFile)
		d.PrivateKey = readFile(t, initialKeyFile)
		return nil
	})
	assert.NoError(t, err)

	updatedCertificate := readFile(t, updatedCertFile)
	updatedKey := readFile(t, updatedKeyFile)

	updatedDomain, err := domainRepository.UpdateDomain(domainName, func(d *acme.Domain) error {
		d.Certificate = updatedCertificate
		d.PrivateKey = updatedKey
		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, domainName, updatedDomain.Name)
	assert.Equal(t, newDomain.ClientID, updatedDomain.ClientID)
	assert.Equal(t, updatedCertificate, updatedDomain.Certificate)
	assert.Equal(t, updatedKey, updatedDomain.PrivateKey)
	assert.NotEqual(t, newDomain.Certificate, updatedDomain.Certificate)
	assert.NotEqual(t, newDomain.PrivateKey, updatedDomain.PrivateKey)
}

func TestUpdateDomain_UpdateFunctionError(t *testing.T) {
	fx := db.NewTestFixture(t)
	defer fx.Close()
	domainRepository := fx.DB.DomainRepository()
	expectedError := errors.New("update failed")
	_, err := domainRepository.UpdateDomain("example.com", func(d *acme.Domain) error {
		return expectedError
	})
	assert.Truef(t, errors.HasCause(err, expectedError), "expected %v to have cause %v", err, expectedError)
}

func readFile(t *testing.T, filename string) []byte {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return bs
}
