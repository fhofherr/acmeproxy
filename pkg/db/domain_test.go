package db_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveNewDomain(t *testing.T) {
	fx := db.NewDBTestFixture(t)
	defer fx.Close()
	domainRepository := fx.DB.DomainRepository()

	domainName := "example.com"
	certFile := filepath.Join("testdata", t.Name(), "certificate.pem")
	keyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
	if *db.FlagUpdate {
		pk := certutil.WritePrivateKeyForTesting(t, keyFile, certutil.EC256, true)
		certutil.WriteCertificateForTesting(t, certFile, domainName, pk, true)
	}
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		t.Fatal(err)
	}
	keyBytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		t.Fatal(err)
	}
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
	assert.Equal(t, domain, actual)
}

// TODO(fhofherr) test update domain
