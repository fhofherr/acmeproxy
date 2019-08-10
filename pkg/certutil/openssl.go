package certutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// CreateOpenSSLPrivateKey creates private key files using OpenSSL.
//
// This is especially useful for testing: in order to test reading key files we
// need such files. Writing them with our own code seems awkward. Therefore we
// use openssl to write those files. The files are checked into version control
// to allow the tests to succeed on systems where openssl is not available.
func CreateOpenSSLPrivateKey(t *testing.T, keyPath string) {
	dir, keyFile := filepath.Split(keyPath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}
	if strings.HasPrefix(keyFile, "ec") {
		createOpenSSLECPrivateKey(t, dir, keyFile)
	} else {
		creteOpenSSLRSAPrivateKey(t, dir, keyFile)
	}
}

func createOpenSSLECPrivateKey(t *testing.T, dir, keyFile string) {
	argv := make([]string, 0, 10)
	switch {
	case strings.HasPrefix(keyFile, "ec256"):
		argv = append(argv, "ecparam", "-name", "prime256v1", "-genkey", "-noout")
	case strings.HasPrefix(keyFile, "ec384"):
		argv = append(argv, "ecparam", "-name", "secp384r1", "-genkey", "-noout")
	default:
		t.Fatalf("don't know how to create key file: %s/%s", dir, keyFile)
	}
	if filepath.Ext(keyFile) == ".pem" {
		argv = append(argv, "-outform", "pem")
	} else {
		argv = append(argv, "-outform", "der")
	}
	argv = append(argv, "-out", filepath.Join(dir, keyFile))
	openssl(t, argv...)
}

func creteOpenSSLRSAPrivateKey(t *testing.T, dir, keyFile string) {
	pemOk := true
	targetFile := filepath.Join(dir, keyFile)
	if filepath.Ext(keyFile) != ".pem" {
		pemOk = false
		targetFile += ".tmp"
	}

	argv := make([]string, 0, 10)
	argv = append(argv, "genrsa", "-out", targetFile)
	switch {
	case strings.HasPrefix(keyFile, "rsa2048"):
		argv = append(argv, "2048")
	case strings.HasPrefix(keyFile, "rsa4096"):
		argv = append(argv, "4096")
	case strings.HasPrefix(keyFile, "rsa8192"):
		argv = append(argv, "8192")
	}
	openssl(t, argv...)
	if !pemOk {
		keyPath := filepath.Join(dir, keyFile)
		openssl(t, "rsa", "-in", targetFile, "-outform", "der", "-out", keyPath)
		err := os.Remove(targetFile)
		if err != nil {
			t.Errorf("failed to remove temp file: %v", err)
		}
	}
}

func openssl(t *testing.T, argv ...string) {
	cmdPath, err := exec.LookPath("openssl")
	if err != nil {
		t.Fatalf("openssl not available: %v", err)
	}
	cmd := exec.Command(cmdPath, argv...)
	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to create keys: %v", err)
	}
}
