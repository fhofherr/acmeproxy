package db

import (
	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/db/internal/dbrecords"
	"github.com/pkg/errors"
)

type domainRepository struct {
	BoltDB     *Bolt
	BucketName string
}

// UpdateDomain updates a domain within the bolt database.
func (d *domainRepository) UpdateDomain(domainName string, f func(d *acme.Domain) error) (acme.Domain, error) {
	var domain acme.Domain

	err := d.BoltDB.updateBucket(d.BucketName, func(b *bucket) error {
		b.readRecord(&dbrecords.BinaryMarshaller{V: domainName}, &dbrecords.BinaryUnmarshaller{V: &domain})
		b.do(func() error {
			return f(&domain)
		})
		b.writeRecord(&dbrecords.BinaryMarshaller{V: domainName}, &dbrecords.BinaryMarshaller{V: &domain})
		return nil
	})
	return domain, errors.Wrap(err, "update domain")
}

// GetDomain reads the domain with the passed domainName from the domain
// repository.
//
// If the domain does not exist the acme.Domain zero value is returned.
func (d *domainRepository) GetDomain(domainName string) (acme.Domain, error) {
	var domain acme.Domain
	err := d.BoltDB.viewBucket(d.BucketName, func(b *bucket) error {
		id := &dbrecords.BinaryMarshaller{V: domainName}
		target := &dbrecords.BinaryUnmarshaller{V: &domain}
		b.readRecord(id, target)
		return nil
	})
	return domain, errors.Wrap(err, "read domain from bucket")
}
