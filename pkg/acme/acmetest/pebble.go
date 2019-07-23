package acmetest

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// Pebble represents an instance of the pebble test server used for testing
// ACME protocol clients.
//
// For more information about pebble see https://github.com/letsencrypt/pebble
type Pebble struct {
	Host           string
	ACMEPort       int
	ManagementPort int
	TestCert       string
	httpClient     *http.Client
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
		t.Skipf("ACMEPROXY_PEBBLE_HOST not set")
	}
	cert := os.Getenv("ACMEPROXY_PEBBLE_TEST_CERT")
	if cert == "" {
		t.Skipf("ACMEPROXY_PEBBLE_TEST_CERT not set")
	}
	if !filepath.IsAbs(cert) {
		t.Fatalf("ACMEPROXY_PEBBLE_TEST_CERT not an absolute path: %s", cert)
	}
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		t.Fatalf("ACMEPROXY_PEBBLE_TEST_CERT does not exist: %s", cert)
	}
	return &Pebble{
		Host:           host,
		ACMEPort:       getPort(t, "ACMEPROXY_PEBBLE_ACME_PORT", 14000),
		ManagementPort: getPort(t, "ACMEPROXY_PEBBLE_MGMT_PORT", 15000),
		TestCert:       cert,
		httpClient:     newHTTPClient(t, cert),
	}
}

// DirectoryURL returns the directory URL of the pebble instance.
func (p *Pebble) DirectoryURL() string {
	return fmt.Sprintf("https://%s:%d/dir", p.Host, p.ACMEPort)
}

// AssertIssuedByPebble asserts that the passed PEM encoded certificate
// is a valid certificate issued by this instance of pebble.
func (p *Pebble) AssertIssuedByPebble(t *testing.T, domain string, certificate []byte) {
	var pebbleCerts []byte
	pebbleCerts = append(pebbleCerts, p.loadCACert(t, "roots")...)
	pebbleCerts = append(pebbleCerts, p.loadCACert(t, "intermediates")...)
	AssertCertificateValid(t, domain, pebbleCerts, certificate)
}

func (p *Pebble) loadCACert(t *testing.T, certType string) []byte {
	certURL := fmt.Sprintf("https://%s:%d/%s/0", p.Host, p.ManagementPort, certType)
	resp, err := p.httpClient.Get(certURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Pebble replied with status %d to GET %s", resp.StatusCode, certURL)
	}
	pemBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	return pemBytes
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

func newHTTPClient(t *testing.T, certFile string) *http.Client {
	certPool := createCertPool(t, certFile)
	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}
	return client
}

func createCertPool(t *testing.T, certFile string) *x509.CertPool {
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		t.Fatal(err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certBytes)
	return certPool
}
