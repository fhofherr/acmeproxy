package db

import (
	"fmt"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/db/internal/dbrecords"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/google/uuid"
)

type userRepository struct {
	BoltDB     *Bolt
	BucketName string
}

// UpdateUser updates a user within the bolt database.
func (r *userRepository) UpdateUser(id uuid.UUID, f func(*acme.User) error) (acme.User, error) {
	const op errors.Op = "db/userRepository.UpdateUser"
	var (
		user acme.User
		err  error
	)

	err = r.BoltDB.updateBucket(r.BucketName, func(b *bucket) error {
		b.readRecord(id, &dbrecords.BinaryUnmarshaller{V: &user})
		b.do(func() error {
			return f(&user)
		})
		b.writeRecord(id, &dbrecords.BinaryMarshaller{V: &user})
		return b.Err
	})
	if err != nil {
		return user, errors.New(op, fmt.Sprintf("user: %v", id), err)
	}
	return user, nil
}

// GetUser obtains a user by its id.
//
// If the user could not be found the acme.User zero value is returned.
func (r *userRepository) GetUser(id uuid.UUID) (acme.User, error) {
	const op errors.Op = "db/userRepository.GetUser"
	var user acme.User

	err := r.BoltDB.viewBucket(r.BucketName, func(b *bucket) error {
		id := &dbrecords.BinaryMarshaller{V: id}
		target := &dbrecords.BinaryUnmarshaller{V: &user}
		b.readRecord(id, target)
		return nil
	})
	return user, errors.Wrap(err, op, fmt.Sprintf("get user: %v", id))
}
