package db_test

import (
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/internal/testsupport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveNewUser(t *testing.T) {
	keyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
	if *testsupport.FlagUpdate {
		certutil.WritePrivateKeyForTesting(t, keyFile, certutil.EC256, true)
	}
	fx := db.NewTestFixture(t)
	defer fx.Close()

	expected := acme.User{
		ID:         uuid.Must(uuid.NewRandom()),
		Key:        certutil.KeyMust(certutil.ReadPrivateKeyFromFile(certutil.EC256, keyFile, true)),
		AccountURL: "https://example.com/some/account",
	}
	userRepository := fx.DB.UserRepository()
	actual, err := userRepository.UpdateUser(expected.ID, func(c *acme.User) error {
		c.ID = expected.ID
		c.Key = expected.Key
		c.AccountURL = expected.AccountURL
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	saved, err := userRepository.GetUser(expected.ID)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
	assert.Equal(t, expected, saved)
}

func TestGetUser(t *testing.T) {
	type test struct {
		name         string
		userID       uuid.UUID
		keyFile      string
		prepare      func(*testing.T, test, *db.TestFixture)
		expectedKind errors.Kind
	}
	var tests = []test{
		{
			name:         "empty repository",
			userID:       uuid.Must(uuid.NewRandom()),
			expectedKind: errors.NotFound,
		},
		{
			name:    "missing user",
			userID:  uuid.Must(uuid.NewRandom()),
			keyFile: "private_key.pem",
			prepare: func(t *testing.T, tt test, fx *db.TestFixture) {
				key := certutil.KeyMust(certutil.ReadPrivateKeyFromFile(certutil.EC256, tt.keyFile, true))
				otherUserID := uuid.Must(uuid.NewRandom())
				repo := fx.DB.UserRepository()
				_, err := repo.UpdateUser(otherUserID, func(c *acme.User) error {
					c.ID = otherUserID
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
				if *testsupport.FlagUpdate {
					certutil.WritePrivateKeyForTesting(t, tt.keyFile, certutil.EC256, true)
				}
			}
			fx := db.NewTestFixture(t)
			defer fx.Close()
			if tt.prepare != nil {
				tt.prepare(t, tt, fx)
			}
			repo := fx.DB.UserRepository()
			_, err := repo.GetUser(tt.userID)
			if err != nil {
				assert.Truef(t, errors.IsKind(err, tt.expectedKind),
					"expected error kind %v, got %v", tt.expectedKind, errors.GetKind(err))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	type updateTest struct {
		name       string
		userID     uuid.UUID
		updatef    func(*acme.User) error
		assertions func(*testing.T, updateTest, *acme.User, error)
	}
	tests := []updateTest{
		{
			name:   "successfully update user",
			userID: uuid.Must(uuid.NewRandom()),
			updatef: func(c *acme.User) error {
				c.AccountURL = "https://example.com/smoe/changed-account"
				return nil
			},
			assertions: func(t *testing.T, tt updateTest, actual *acme.User, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "https://example.com/smoe/changed-account", actual.AccountURL)
				assert.Equal(t, tt.userID, actual.ID)
			},
		},
		{
			name:   "update function returns error",
			userID: uuid.Must(uuid.NewRandom()),
			updatef: func(*acme.User) error {
				return errors.New("update function failed")
			},
			assertions: func(t *testing.T, tt updateTest, _ *acme.User, err error) {
				cause := errors.New("update function failed")
				assert.Truef(t, errors.Is(err, cause), "expected %v to have cause %v", err, cause)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			initialKeyFile := filepath.Join("testdata", t.Name(), "private_key.pem")
			if *testsupport.FlagUpdate {
				certutil.WritePrivateKeyForTesting(t, initialKeyFile, certutil.EC256, true)
			}
			fx := db.NewTestFixture(t)
			defer fx.Close()
			userRepository := fx.DB.UserRepository()

			key := certutil.KeyMust(certutil.ReadPrivateKeyFromFile(certutil.EC256, initialKeyFile, true))
			_, err := userRepository.UpdateUser(tt.userID, func(c *acme.User) error {
				c.ID = tt.userID
				c.Key = key
				c.AccountURL = "https://example.com/some/new-account"
				return nil
			})
			assert.NoError(t, err)

			actual, err := userRepository.UpdateUser(tt.userID, tt.updatef)
			tt.assertions(t, tt, &actual, err)
		})
	}
}
