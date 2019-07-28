package acme

import "github.com/fhofherr/acmeproxy/pkg/acme/internal/challenge"

type Config struct {
	DirectoryURL string
}

// NewClient creates a new ACME client.
func NewClient(cfg Config) *Client {
	return &Client{
		DirectoryURL: cfg.DirectoryURL,
		HTTP01Solver: challenge.NewHTTP01Solver(),
	}
}
