package db

import (
	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/db/internal/dbrecords"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

type clientRepository struct {
	BoltDB     *Bolt
	BucketName string
}

// UpdateClient updates a client within the bolt database.
func (r *clientRepository) UpdateClient(id uuid.UUID, f func(*acme.Client) error) (acme.Client, error) {
	var (
		client  acme.Client
		idBytes []byte
		err     error
	)

	idBytes, err = dbrecords.Marshal(id)
	if err != nil {
		return client, errors.Wrap(err, "marshal client id")
	}

	err = r.BoltDB.updateBucket(r.BucketName, func(bucket *bbolt.Bucket) error {
		err = r.BoltDB.readRecord(bucket, idBytes, &client)
		if err != nil {
			return errors.Wrap(err, "read current client record")
		}
		err = f(&client)
		if err != nil {
			return errors.Wrap(err, "apply update func to client")
		}
		bs, err := dbrecords.Marshal(&client)
		if err != nil {
			return errors.Wrapf(err, "marshal client")
		}
		return errors.Wrapf(bucket.Put(idBytes, bs), "save client: %v", id)
	})
	return client, errors.Wrapf(err, "update client: %v", id)
}

// GetClient obtains a client by its id.
//
// If the client could not be found the acme.Client zero value is returned.
func (r *clientRepository) GetClient(id uuid.UUID) (acme.Client, error) {
	var client acme.Client

	idBytes, err := dbrecords.Marshal(id)
	if err != nil {
		return client, errors.Wrap(err, "marshal client id")
	}
	err = r.BoltDB.viewBucket(r.BucketName, func(bucket *bbolt.Bucket) error {
		return r.BoltDB.readRecord(bucket, idBytes, &client)
	})
	return client, errors.Wrapf(err, "get client: %v", id)
}
