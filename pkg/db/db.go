package db

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

// DefaultDBFileMode is the default file mode a bolt database file should
// be created with if nothing else is specified.
const DefaultDBFileMode = 0600

// Bolt represents the bbolt database used by acmeproxy to store its data.
//
// Bolt manages the data file used by bbolt and provides factory methods
// for the various repositories used throughout acmeproxy.
//
// See https://github.com/etcd-io/bbolt for more information about bbolt.
type Bolt struct {
	FilePath string
	FileMode os.FileMode
	db       *bbolt.DB
	mu       sync.Mutex
}

// Open opens the bolt database.
func (b *Bolt) Open() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.db != nil {
		return nil
	}
	if b.FileMode == 0 {
		b.FileMode = DefaultDBFileMode
	}
	dbDir := filepath.Dir(b.FilePath)
	err := os.MkdirAll(dbDir, 0755)
	if err != nil {
		return errors.Wrap(err, "create db directory")
	}
	db, err := bbolt.Open(b.FilePath, b.FileMode, nil)
	if err != nil {
		return errors.Wrap(err, "open bolt database")
	}
	b.db = db
	return nil
}

// Close closes the bolt database.
func (b *Bolt) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.db == nil {
		return nil
	}
	return errors.Wrap(b.db.Close(), "close bolt database")
}

// ClientRepository returns an instance of a client repository.
func (b *Bolt) ClientRepository() acme.ClientRepository {
	return &clientRepository{
		BoltDB:     b,
		BucketName: "clients",
	}
}

// DomainRepository returns an instance of a domain repository.
func (b *Bolt) DomainRepository() acme.DomainRepository {
	return &domainRepository{
		BoltDB:     b,
		BucketName: "domains",
	}
}

func (b *Bolt) updateBucket(name string, update func(*bbolt.Bucket) error) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return errors.Wrapf(err, "create bucket: %s", name)
		}
		return update(bucket)
	})
}
