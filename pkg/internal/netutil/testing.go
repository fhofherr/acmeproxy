package netutil

import (
	"testing"
	"time"
)

// GetAddr reads a servers address from addrC.
func GetAddr(t *testing.T, addrC <-chan string) string {
	select {
	case addr := <-addrC:
		return addr
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Time-out after 10ms")
	}
	return ""
}

// GetErr reads an error from errC.
func GetErr(t *testing.T, errC <-chan error) error {
	select {
	case err := <-errC:
		return err
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Time-out after 10ms")
	}
	return nil
}
