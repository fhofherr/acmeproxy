package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fhofherr/acmeproxy/pkg/acme"
	"github.com/fhofherr/acmeproxy/pkg/errors"
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
	const op errors.Op = "db/bolt.Open"

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
		return errors.New(op, "create directory", err)
	}
	db, err := bbolt.Open(b.FilePath, b.FileMode, nil)
	if err != nil {
		return errors.New(op, "open database", err)
	}
	b.db = db
	return nil
}

// Close closes the bolt database.
func (b *Bolt) Close() error {
	const op errors.Op = "db/bolt.Close"

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.db == nil {
		return nil
	}
	err := b.db.Close()
	if err != nil {
		return errors.New(op, "close database", err)
	}
	return nil
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

func (b *Bolt) viewBucket(name string, view func(*bucket) error) error {
	const op errors.Op = "db/bolt.viewBucket"

	return b.db.View(func(tx *bbolt.Tx) error {
		boltBucket := tx.Bucket([]byte(name))
		if boltBucket == nil {
			return errors.New(op, errors.NotFound, fmt.Sprintf("bucket: %s", name))
		}
		bucket := &bucket{Bkt: boltBucket}
		err := view(bucket)
		if err != nil {
			return err
		}
		return bucket.Err
	})
}

func (b *Bolt) updateBucket(name string, update func(*bucket) error) error {
	const op errors.Op = "db/bolt.updateBucket"

	return b.db.Update(func(tx *bbolt.Tx) error {
		boltBucket, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return errors.New(op, "create bucket", err)
		}
		bucket := &bucket{Bkt: boltBucket}
		err = update(bucket)
		if err != nil {
			return err
		}
		return bucket.Err
	})
}
