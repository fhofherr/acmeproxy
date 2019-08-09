package certutil

// KeyType represents the types of cryptographic keys supported by acmeproxy.
//
// The supported key types are dictated by what our ACME client library
// supports.
type KeyType int

const (
	// EC256 represents an ECDSA key using an elliptic curve implementing P-256.
	EC256 KeyType = iota
	// EC384 represents an ECDSA key using an elliptic curve implementing P-384.
	EC384
	// RSA2048 represents an RSA key with a size of 2048 bits.
	RSA2048
	// RSA4096 represents an RSA key with a size of 4096 bits.
	RSA4096
	// RSA8192 represents an RSA key with a size of 8192 bits.
	RSA8192
)
