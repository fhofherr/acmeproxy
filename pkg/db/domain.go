package db

import (
	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/db/internal/dbrecords"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

type domainRepository struct {
	BoltDB     *Bolt
	BucketName string
}

// UpdateDomain updates a domain within the bolt database.
func (d *domainRepository) UpdateDomain(domainName string, f func(d *acme.Domain) error) (acme.Domain, error) {
	var (
		domain acme.Domain
	)
	err := d.BoltDB.updateBucket(d.BucketName, func(bucket *bbolt.Bucket) error {
		err := f(&domain)
		if err != nil {
			return errors.Wrap(err, "apply update func to domain")
		}
		bs, err := dbrecords.Marshal(&domain)
		if err != nil {
			return errors.Wrapf(err, "marshal client")
		}
		return errors.Wrapf(bucket.Put([]byte(domainName), bs), "save domain: %v", domainName)
	})
	return domain, errors.Wrap(err, "update domain")
}

func (d *domainRepository) GetDomain(string) (acme.Domain, error) {
	panic("implement me")
}
