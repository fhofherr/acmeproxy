package acmeclient

import (
	"fmt"
	"testing"

	"github.com/fhofherr/acmeproxy/pkg/certutil"
	"github.com/fhofherr/golf/log"
	"github.com/go-acme/lego/certcrypto"
	"github.com/stretchr/testify/assert"
)

func TestLoggerAdapter(t *testing.T) {
	tests := []struct {
		name    string
		logCall func(*loggerAdapter)
		level   string
		message string
	}{
		{
			name: "print",
			logCall: func(a *loggerAdapter) {
				a.Print("some", "message")
			},
			level:   "info",
			message: fmt.Sprint("some", "message"),
		},
		{
			// We do not want a newline. Thus the behavior should be the same
			// as for print.
			name: "println",
			logCall: func(a *loggerAdapter) {
				a.Println("some", "message")
			},
			level:   "info",
			message: fmt.Sprint("some", "message"),
		},
		{
			name: "printf",
			logCall: func(a *loggerAdapter) {
				a.Printf("%s--%s", "some", "message")
			},
			level:   "info",
			message: fmt.Sprintf("%s--%s", "some", "message"),
		},
		{
			name: "fatal",
			logCall: func(a *loggerAdapter) {
				a.Fatal("some", "message")
			},
			level:   "error",
			message: fmt.Sprint("some", "message"),
		},
		{
			// We do not want a newline. Thus the behavior should be the same
			// as for fatal.
			name: "fatalln",
			logCall: func(a *loggerAdapter) {
				a.Fatalln("some", "message")
			},
			level:   "error",
			message: fmt.Sprint("some", "message"),
		},
		{
			name: "fatalf",
			logCall: func(a *loggerAdapter) {
				a.Fatalf("%s--%s", "some", "message")
			},
			level:   "error",
			message: fmt.Sprintf("%s--%s", "some", "message"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testLogger := &log.TestLogger{}
			adapter := &loggerAdapter{
				Logger: testLogger,
			}
			tt.logCall(adapter)
			testLogger.AssertHasMatchingLogEntries(t, 1, func(e log.TestLogEntry) bool {
				return e["level"] == tt.level && e["message"] == tt.message
			})
		})
	}
}

func TestKeyTypeToLegoKeyType(t *testing.T) {
	tests := []struct {
		name        string
		keyType     certutil.KeyType
		legoKeyType certcrypto.KeyType
	}{
		{"EC256", certutil.EC256, certcrypto.EC256},
		{"EC384", certutil.EC384, certcrypto.EC384},
		{"RSA2048", certutil.RSA2048, certcrypto.RSA2048},
		{"RSA4096", certutil.RSA4096, certcrypto.RSA4096},
		{"RSA8192", certutil.RSA8192, certcrypto.RSA8192},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := legoKeyType(tt.keyType)
			assert.NoError(t, err)
			assert.Equal(t, tt.legoKeyType, actual)
		})
	}
}

func TestInvalidKeyTypeToLegoKeyType(t *testing.T) {
	_, err := legoKeyType(certutil.KeyType(-1))
	assert.Error(t, err)
}
