package testsupport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
)

var (
	pebbleDir = os.Getenv("ACMEPROXY_PEBBLE_DIR")
)

// Pebble represents an instance of the pebble test server used for testing
// ACME protocol clients.
//
// For more information about pebble see https://github.com/letsencrypt/pebble
type Pebble struct {
	TestCert        string
	configJSON      string
	config          pebbleConfig
	pebbleDir       string
	pebbleCMD       *exec.Cmd
	challtestsrvCMD *exec.Cmd
	httpClient      *http.Client
}

// SkipIfPebbleDisabled checks whether the test should be skipped because
// pebble is not available.
//
// At least the ACMEPROXY_PEBBLE_DIR environment variable has to be set,
// otherwise the test is skipped.
func SkipIfPebbleDisabled(t *testing.T) bool {
	if pebbleDir == "" {
		t.Skip("Pebble disabled. Set ACMEPROXY_PEBBLE_DIR to enable.")
		return true
	}
	return false
}

// NewPebble creates a new Pebble instance by reading the necessary
// configuration settings from environment variables or using defaults.
func NewPebble(t *testing.T, configJSON, dnsPort string) *Pebble {
	if pebbleDir == "" {
		t.Fatal("ACMEPROXY_PEBBLE_DIR not set")
	}

	pebble := findPebbleCommand(t, "pebble")
	pebbleCMD := exec.Command(
		pebble,
		"-strict",
		"-config",
		configJSON,
		"-dnsserver",
		"127.0.0.1:"+dnsPort)
	pebbleCMD.Env = []string{"PEBBLE_VA_NOSLEEP=1"}

	challtestsrv := findPebbleCommand(t, "pebble-challtestsrv")
	challtestsrvCMD := exec.Command(
		challtestsrv,
		"-defaultIPv6", "",
		"-defaultIPv4", "127.0.0.1",
		"-dns01", ":"+dnsPort,
		"-http01", "",
		"-https01", "")
	pebbleConfig := readPebbleConfig(t, configJSON)

	pebbleCert := filepath.Join(pebbleDir, "test", "certs", "pebble.minica.pem")
	if _, err := os.Stat(pebbleCert); os.IsNotExist(err) {
		t.Fatalf("Pebble certificate does not exist: %s", pebbleCert)
	}
	return &Pebble{
		TestCert:        pebbleCert,
		configJSON:      configJSON,
		config:          pebbleConfig,
		pebbleDir:       pebbleDir,
		pebbleCMD:       pebbleCMD,
		challtestsrvCMD: challtestsrvCMD,
		httpClient:      newHTTPClient(t, pebbleCert),
	}
}

// Start starts the pebble server for the test.
func (p *Pebble) Start(t *testing.T) {
	if err := p.challtestsrvCMD.Start(); err != nil {
		t.Fatalf("Failed to start pebble-challtestsrv: %v", err)
	}
	if err := p.pebbleCMD.Start(); err != nil {
		t.Fatalf("Failed to start pebble: %v", err)
	}
	p.WaitReady(t)
}

// Stop stops the pebble server after the test.
func (p *Pebble) Stop(t *testing.T) {
	stopCMD(t, p.pebbleCMD)
	stopCMD(t, p.challtestsrvCMD)
}

// WaitReady waits for pebble to become ready.
func (p *Pebble) WaitReady(t *testing.T) {
	if p.pebbleCMD.Process == nil {
		t.Fatal("Pebble not started")
	}
	url := p.DirectoryURL()
	Retry(t, 10, 10*time.Millisecond, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		t.Log("Checking pebble readiness")
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := p.httpClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})
}

func stopCMD(t *testing.T, cmd *exec.Cmd) {
	if cmd.Process == nil {
		t.Logf("Not started: %v", cmd)
		return
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Errorf("Failed to kill: %v", cmd)
	}
	cmd.Wait() // nolint: errcheck
}

// HTTPPort returns the port pebble uses to solve HTTP01 challenges.
func (p *Pebble) HTTPPort() int {
	return p.config.Pebble.HTTPPort
}

// DirectoryURL returns the directory URL of the pebble instance.
func (p *Pebble) DirectoryURL() string {
	return fmt.Sprintf("https://%s/dir", p.config.Pebble.ListenAddress)
}

// AccountURLPrefix returns the prefix of all account URLs belonging to
// accounts issued by this instance of pebble.
func (p *Pebble) AccountURLPrefix() string {
	return fmt.Sprintf("https://%s/my-account", p.config.Pebble.ListenAddress)
}

// AssertIssuedByPebble asserts that the passed PEM encoded certificate
// is a valid certificate issued by this instance of pebble.
func (p *Pebble) AssertIssuedByPebble(t *testing.T, domain string, certificate []byte) {
	var pebbleCerts []byte
	pebbleCerts = append(pebbleCerts, p.loadCACert(t, "roots")...)
	pebbleCerts = append(pebbleCerts, p.loadCACert(t, "intermediates")...)
	certutil.AssertCertificateValid(t, domain, pebbleCerts, certificate)
}

func (p *Pebble) loadCACert(t *testing.T, certType string) []byte {
	certURL := fmt.Sprintf("https://%s/%s/0", p.config.Pebble.ManagementListenAddress, certType)
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

func findPebbleCommand(t *testing.T, name string) string {
	cmd := filepath.Join(pebbleDir, name)
	_, err := os.Stat(cmd)
	if os.IsNotExist(err) {
		t.Fatalf("Command does not exist: %s", cmd)
	}
	return cmd
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
		Timeout:   time.Second * 2,
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

func readPebbleConfig(t *testing.T, configJSON string) pebbleConfig {
	data, err := ioutil.ReadFile(configJSON)
	if err != nil {
		t.Fatalf("Could not read %s: %v", configJSON, err)
	}
	var cfg pebbleConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Could not unmarshal %s: %v", configJSON, err)
	}
	return cfg
}

type pebbleConfig struct {
	Pebble struct {
		ListenAddress           string
		ManagementListenAddress string
		HTTPPort                int
		TLSPort                 int
		Certificate             string
		PrivateKey              string
		OCSPResponderURL        string
	}
}
