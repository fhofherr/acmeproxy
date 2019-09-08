package db

import (
	"encoding"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CreateTmpDir creates a temporary directory.
//
// It returns the path to the created temporary directory an a clean-up function
// which allows to delete it.
func CreateTmpDir(t *testing.T) (string, func()) {
	prefix := strings.Replace(t.Name(), string(os.PathSeparator), "_", -1)
	tmpDir, err := ioutil.TempDir("", prefix)
	if err != nil {
		t.Fatal(err)
	}
	return tmpDir, func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Error(err)
		}
	}
}

// TestFixture wraps a database suitable for testing.
//
// After the test is done the fixture needs to be discarded by calling
// its Close method.
//
// Use NewTestFixture to create an instance of this type.
type TestFixture struct {
	DB           *Bolt
	tmpDir       string
	deleteTmpDir func()
	t            *testing.T
}

// NewTestFixture creates a new database test fixture.
func NewTestFixture(t *testing.T) *TestFixture {
	tmpDir, deleteTmpDir := CreateTmpDir(t)
	dbPath := filepath.Join(tmpDir, "bolt.db")
	boltDB := &Bolt{
		FilePath: dbPath,
		FileMode: 0600,
	}
	err := boltDB.Open()
	if err != nil {
		t.Fatal(err)
	}
	return &TestFixture{
		DB:           boltDB,
		tmpDir:       tmpDir,
		deleteTmpDir: deleteTmpDir,
		t:            t,
	}
}

// Close discards the database test fixture and gets rid of all files
// and directories created for the test database.
func (fx *TestFixture) Close() error {
	if err := fx.DB.Close(); err != nil {
		fx.t.Error(err)
	}
	fx.deleteTmpDir()
	return nil
}

// CreateBucket adds an empty bucket with the passed name to the database.
// It fails the test if an error occurs.
func (fx *TestFixture) CreateBucket(name string) {
	err := fx.DB.updateBucket(name, func(b *bucket) error {
		return nil
	})
	if err != nil {
		fx.t.Fatalf("TestFixture.CreateBucket: %v", err)
	}
}

// CreateBucketWithKey add a bucket with the given name to the database. The
// bucket contains the passed key with the passed value.
func (fx *TestFixture) CreateBucketWithKey(name string, key encoding.BinaryMarshaler, value encoding.BinaryMarshaler) {
	err := fx.DB.updateBucket(name, func(b *bucket) error {
		b.writeRecord(key, value)
		return nil
	})
	if err != nil {
		fx.t.Fatalf("TestFixture.CreateBucketWithKey: %v", err)
	}
}
