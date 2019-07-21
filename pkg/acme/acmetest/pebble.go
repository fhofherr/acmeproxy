package acmetest

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

// Pebble represents an instance of the pebble test server used for testing
// ACME protocol clients.
//
// For more information about pebble see https://github.com/letsencrypt/pebble
type Pebble struct {
	Host           string
	ACMEPort       int
	ManagementPort int
}

// NewPebble creates a new Pebble instance by reading the necessary
// configuration settings from environment variables or using defaults.
//
// At least the ACMEPROXY_PEBBLE_HOST environment variable has to be set,
// otherwise the test is skipped. Additionally the ACMEPort may be set using
// the ACMEPROXY_PEBBLE_ACME_PORT environment variable, and the management port
// by using the ACMEPROXY_PEBBLE_MGMT_PORT environment variable.
func NewPebble(t *testing.T) *Pebble {
	host := os.Getenv("ACMEPROXY_PEBBLE_HOST")
	if host == "" {
		t.Skipf("No ACMEPROXY_PEBBLE_HOST configured")
	}
	return &Pebble{
		Host:           host,
		ACMEPort:       getPort(t, "ACMEPROXY_PEBBLE_ACME_PORT", 14000),
		ManagementPort: getPort(t, "ACMEPROXY_PEBBLE_MGMT_PORT", 15000),
	}
}

// DirectoryURL returns the directory URL of the pebble instance.
func (p *Pebble) DirectoryURL() string {
	return fmt.Sprintf("https://%s:%d/dir", p.Host, p.ACMEPort)
}

func getPort(t *testing.T, key string, defaultValue int) int {
	portStr := os.Getenv(key)
	if portStr == "" {
		return defaultValue
	}
	portNo, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Failed to get port from %s: %v", key, err)
	}
	return portNo
}
