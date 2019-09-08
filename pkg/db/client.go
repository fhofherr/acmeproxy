package db

import (
	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/db/internal/dbrecords"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type clientRepository struct {
	BoltDB     *Bolt
	BucketName string
}

// UpdateClient updates a client within the bolt database.
func (r *clientRepository) UpdateClient(id uuid.UUID, f func(*acme.Client) error) (acme.Client, error) {
	var (
		client acme.Client
		err    error
	)

	err = r.BoltDB.updateBucket(r.BucketName, func(b *bucket) error {
		b.readRecord(id, &dbrecords.BinaryUnmarshaller{V: &client})
		b.do(func() error {
			return f(&client)
		})
		b.writeRecord(id, &dbrecords.BinaryMarshaller{V: &client})
		return b.Err
	})
	return client, errors.Wrapf(err, "update client: %v", id)
}

// GetClient obtains a client by its id.
//
// If the client could not be found the acme.Client zero value is returned.
func (r *clientRepository) GetClient(id uuid.UUID) (acme.Client, error) {
	var client acme.Client

	err := r.BoltDB.viewBucket(r.BucketName, func(b *bucket) error {
		id := &dbrecords.BinaryMarshaller{V: id}
		target := &dbrecords.BinaryUnmarshaller{V: &client}
		b.readRecord(id, target)
		return nil
	})

	return client, errors.Wrapf(err, "get client: %v", id)
}
