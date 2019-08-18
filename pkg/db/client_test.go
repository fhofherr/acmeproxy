package db_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveNewClientToBoltDB(t *testing.T) {
	tmpDir, tearDown := createTmpDir(t)
	defer tearDown()

	bolt := db.Bolt{
		FilePath: filepath.Join(tmpDir, "test.db"),
	}
	err := bolt.Open()
	if !assert.NoError(t, err) {
		return
	}
	defer bolt.Close()

	expected := acme.Client{
		ID:         uuid.Must(uuid.NewRandom()),
		Key:        newPrivateKey(t),
		AccountURL: "https://example.com/some/account",
	}
	clientRepository := bolt.ClientRepository()
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

func newPrivateKey(t *testing.T) crypto.PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}
