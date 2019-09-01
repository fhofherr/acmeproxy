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

func TestSaveNewClient(t *testing.T) {
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

func TestUpdateClient(t *testing.T) {
	initialKeyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
	if *db.FlagUpdate {
		certutil.WritePrivateKeyForTesting(t, initialKeyFile, certutil.EC256, true)
	}
	fx := db.NewDBTestFixture(t)
	defer fx.Close()

	clientRepository := fx.DB.ClientRepository()

	clientID := uuid.Must(uuid.NewRandom())
	initialURL := "https://example.com/some/new-account"
	key := certutil.KeyMust(certutil.ReadPrivateKeyFromFile(certutil.EC256, initialKeyFile, true))
	_, err := clientRepository.UpdateClient(clientID, func(c *acme.Client) error {
		c.ID = clientID
		c.Key = key
		c.AccountURL = initialURL
		return nil
	})
	assert.NoError(t, err)

	changedURL := "https://example.com/smoe/changed-account"
	actual, err := clientRepository.UpdateClient(clientID, func(c *acme.Client) error {
		c.AccountURL = changedURL
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, changedURL, actual.AccountURL)
	assert.Equal(t, clientID, actual.ID)
	assert.Equal(t, key, actual.Key)
}
