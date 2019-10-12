package db_test

import (
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveNewClient(t *testing.T) {
	keyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
	if *db.FlagUpdate {
		certutil.WritePrivateKeyForTesting(t, keyFile, certutil.EC256, true)
	}
	fx := db.NewTestFixture(t)
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
	assert.Equal(t, expected, actual)
	assert.Equal(t, expected, saved)
}

func TestGetClient(t *testing.T) {
	type test struct {
		name         string
		clientID     uuid.UUID
		keyFile      string
		prepare      func(*testing.T, test, *db.TestFixture)
		expectedKind errors.Kind
	}
	var tests = []test{
		{
			name:         "empty repository",
			clientID:     uuid.Must(uuid.NewRandom()),
			expectedKind: errors.NotFound,
		},
		{
			name:     "missing client",
			clientID: uuid.Must(uuid.NewRandom()),
			keyFile:  filepath.Join("testdata", t.Name(), "private_key.pem"),
			prepare: func(t *testing.T, tt test, fx *db.TestFixture) {
				key := certutil.KeyMust(certutil.ReadPrivateKeyFromFile(certutil.EC256, tt.keyFile, true))
				otherClientID := uuid.Must(uuid.NewRandom())
				repo := fx.DB.ClientRepository()
				_, err := repo.UpdateClient(otherClientID, func(c *acme.Client) error {
					c.ID = otherClientID
					c.Key = key
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			expectedKind: errors.NotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.keyFile != "" {
				tt.keyFile = filepath.Join("testdata", t.Name(), tt.keyFile)
				if *db.FlagUpdate {
					certutil.WritePrivateKeyForTesting(t, tt.keyFile, certutil.EC256, true)
				}
			}
			fx := db.NewTestFixture(t)
			defer fx.Close()
			if tt.prepare != nil {
				tt.prepare(t, tt, fx)
			}
			repo := fx.DB.ClientRepository()
			_, err := repo.GetClient(tt.clientID)
			if err != nil {
				assert.Truef(t, errors.IsKind(err, tt.expectedKind),
					"expected error kind %v, got %v", tt.expectedKind, errors.GetKind(err))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateClient(t *testing.T) {
	type updateTest struct {
		name       string
		clientID   uuid.UUID
		updatef    func(*acme.Client) error
		assertions func(*testing.T, updateTest, *acme.Client, error)
	}
	tests := []updateTest{
		{
			name:     "successfully update client",
			clientID: uuid.Must(uuid.NewRandom()),
			updatef: func(c *acme.Client) error {
				c.AccountURL = "https://example.com/smoe/changed-account"
				return nil
			},
			assertions: func(t *testing.T, tt updateTest, actual *acme.Client, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "https://example.com/smoe/changed-account", actual.AccountURL)
				assert.Equal(t, tt.clientID, actual.ID)
			},
		},
		{
			name:     "update function returns error",
			clientID: uuid.Must(uuid.NewRandom()),
			updatef: func(*acme.Client) error {
				return errors.New("update function failed")
			},
			assertions: func(t *testing.T, tt updateTest, _ *acme.Client, err error) {
				cause := errors.New("update function failed")
				assert.Truef(t, errors.HasCause(err, cause), "expected %v to have cause %v", err, cause)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			initialKeyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
			if *db.FlagUpdate {
				certutil.WritePrivateKeyForTesting(t, initialKeyFile, certutil.EC256, true)
			}
			fx := db.NewTestFixture(t)
			defer fx.Close()
			clientRepository := fx.DB.ClientRepository()

			key := certutil.KeyMust(certutil.ReadPrivateKeyFromFile(certutil.EC256, initialKeyFile, true))
			_, err := clientRepository.UpdateClient(tt.clientID, func(c *acme.Client) error {
				c.ID = tt.clientID
				c.Key = key
				c.AccountURL = "https://example.com/some/new-account"
				return nil
			})
			assert.NoError(t, err)

			actual, err := clientRepository.UpdateClient(tt.clientID, tt.updatef)
			tt.assertions(t, tt, &actual, err)
		})
	}
}
