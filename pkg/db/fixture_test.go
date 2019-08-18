package db

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type DBTestFixture struct {
	tmpDir string
	DB     *Bolt
	t      *testing.T
}

func NewDBTestFixture(t *testing.T) *DBTestFixture {
	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "bolt.db")
	boltDB := &Bolt{
		FilePath: dbPath,
		FileMode: 0600,
	}
	err = boltDB.Open()
	if err != nil {
		t.Fatal(err)
	}
	return &DBTestFixture{
		tmpDir: tmpDir,
		DB:     boltDB,
		t:      t,
	}
}

func (fx *DBTestFixture) Close() error {
	if err := fx.DB.Close(); err != nil {
		fx.t.Error(err)
	}
	if err := os.RemoveAll(fx.tmpDir); err != nil {
		fx.t.Error(err)
	}
	return nil
}
