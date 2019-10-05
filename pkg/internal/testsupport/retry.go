package testsupport

import (
	"testing"
	"time"
)

// Retry retries op at most n times before it fails t. It sleeps for
// d after the first failure and then increases d by the next power of itself.
func Retry(t *testing.T, n int, d time.Duration, op func() error) {
	var err error

	for i := 0; i < n; i++ {
		err = op()
		if err == nil {
			return
		}
		t.Logf("Retry: attempt %d of %d failed.", i+1, n)
		time.Sleep(d)
		d *= d
	}
	t.Fatal(err)
}
