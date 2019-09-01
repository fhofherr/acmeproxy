package db_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestCreateNewBoltDB(t *testing.T) {
	tmpDir, tearDown := createTmpDir(t)
	defer tearDown()

	tests := []struct {
		name     string
		filePath string
		fileMode os.FileMode
	}{
		{
			name:     "create db file with default mode",
			filePath: filepath.Join(tmpDir, "default_mode", "test.db"),
		},
		{
			name:     "create db file with explicit mode",
			filePath: filepath.Join(tmpDir, "explicit_mode", "test.db"),
			fileMode: 0755,
		},
	}

	for _, tt := range tests {
		boltDB := &db.Bolt{
			FilePath: tt.filePath,
			FileMode: tt.fileMode,
		}
		err := boltDB.Open()
		if !assert.NoError(t, err) || !assert.FileExists(t, boltDB.FilePath) {
			return
		}
		expectedMode := tt.fileMode
		if expectedMode == 0 {
			expectedMode = db.DefaultDBFileMode
		}
		fileInfo, err := os.Stat(tt.filePath)
		assert.NoError(t, err)
		assert.Equal(t, expectedMode, fileInfo.Mode())

		err = boltDB.Close()
		assert.NoError(t, err)
	}
}

func TestTargetDirectoryNotWritable(t *testing.T) {
	tmpDir, tearDown := createTmpDir(t)
	defer tearDown()
	notWritable := filepath.Join(tmpDir, "not_writable")
	err := os.Mkdir(notWritable, 0500)
	if !assert.NoError(t, err) {
		return
	}

	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "can't create db file",
			filePath: filepath.Join(notWritable, "test.db"),
		},
		{
			name:     "can't create intermediate directories",
			filePath: filepath.Join(notWritable, "intermediate", "test.db"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			boltDB := db.Bolt{
				FilePath: tt.filePath,
			}
			err = boltDB.Open()
			assert.Error(t, err)
		})
	}
}

func TestOpenDBTwice(t *testing.T) {
	tmpDir, tearDown := createTmpDir(t)
	defer tearDown()
	boltDB := db.Bolt{FilePath: filepath.Join(tmpDir, "test.db")}
	assert.NoError(t, boltDB.Open())
	assert.NoError(t, boltDB.Open())
}

func TestCloseNonOpenDB(t *testing.T) {
	tmpDir, tearDown := createTmpDir(t)
	defer tearDown()
	boltDB := db.Bolt{FilePath: filepath.Join(tmpDir, "test.db")}
	assert.NoError(t, boltDB.Close())
}

func createTmpDir(t *testing.T) (string, func()) {
	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	return tmpDir, func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Error(err)
		}
	}
}
