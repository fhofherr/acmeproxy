package testsupport

import (
	"os"
	"testing"
)

const legoCACertificatesKey = "LEGO_CA_CERTIFICATES"

// SetLegoCACertificates sets the LEGO_CA_CERTIFICATES environment variable to
// certPath. This makes all new instances of the lego ACME client use the
// certificate located at certPath when connecting to the ACME server.
//
// SetLegoCACertificates returns a function which reverts any changes this
// made to the environment. Always use it in the following way:
//
//     reset := SetLegoCACertificates(t, somePath)
//     defer reset()
//     ... test code ...
func SetLegoCACertificates(t *testing.T, certPath string) func() {
	cur, set := os.LookupEnv(legoCACertificatesKey)
	if err := os.Setenv(legoCACertificatesKey, certPath); err != nil {
		t.Fatalf("Could not set %s to %s: %v", legoCACertificatesKey, certPath, err)
	}
	return func() {
		if !set {
			return
		}
		if err := os.Setenv(legoCACertificatesKey, cur); err != nil {
			t.Errorf("Could not restore %s to %s: %v", legoCACertificatesKey, cur, err)
		}
	}
}
