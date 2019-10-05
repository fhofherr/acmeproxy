package testsupport

import (
	"io/ioutil"
	"os"
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
