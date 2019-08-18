package db_test

import (
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveNewClientToBoltDB(t *testing.T) {
	keyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
	if *db.FlagUpdate {
		certutil.WritePrivateKeyForTesting(t, keyFile, certutil.EC256, true)
	}
	fx := db.NewDBTestFixture(t)
	defer fx.Close()

	expected := acme.Client{
		ID:         uuid.Must(uuid.NewRandom()),
		Key:        certutil.KeyMust(certutil.ReadPrivateKeyFromFile(certutil.EC256, keyFile, true)),
		AccountURL: "https://example.com/some/account",
	}
	clientRepository := fx.DB.ClientRepository()
	actual, err := clientRepository.UpdateClient(expected.ID, func(c *acme.Client) error {
		c.ID = expected.ID
		c.Key = expected.Key
		c.AccountURL = expected.AccountURL
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	saved, err := clientRepository.GetClient(expected.ID)
	assert.NoError(t, err)
	assert.Equal(t, expected, saved)
}

// TODO(fhofherr) test update client
