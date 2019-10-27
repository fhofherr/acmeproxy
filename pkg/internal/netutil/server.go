package netutil

import (
	"net"
	"net/http"

	"github.com/fhofherr/acmeproxy/pkg/errors"
)

// Server wraps the Serve method found on the http.Server and grpc.Server
// types.
type Server interface {
	Serve(net.Listener) error
}

type netutilOptions struct {
	Addr     string
	AddrChan chan<- string
	Listener net.Listener
}

// Option represents a configurable option for the helpers in netutil.
type Option func(*netutilOptions)

// WithAddr sets the network address the server started by ListenAndServe will
// listen on.
func WithAddr(addr string) Option {
	return func(opts *netutilOptions) {
		opts.Addr = addr
	}
}

// NotifyAddr passes a channel to ListenAndServe which is used to send the
// address the Server is listening on to the caller.
func NotifyAddr(c chan<- string) Option {
	return func(opts *netutilOptions) {
		opts.AddrChan = c
	}
}

// WithListener tells ListenAndServe to use the passed listener instead of
// creating a new one.
func WithListener(l net.Listener) Option {
	return func(opts *netutilOptions) {
		opts.Listener = l
	}
}

func applyOptions(nuOpts *netutilOptions, options []Option) *netutilOptions {
	for _, opt := range options {
		opt(nuOpts)
	}
	return nuOpts
}

// ListenAndServe creates a net.Listener and for addr and calls s.Serve with
// it.
//
// Its behavior is very similar to net/http.ListenAndServe. In contrast to
// net/http.ListenAndServe it does not return net/http.ErrServerClosed when the
// s was closed. Instead it returns nil. Errors are only returned if an error
// occurred.
//
// Callers may pass any of the following options to ListenAndServe:
//
//    * NotifyAddr
//    * WithAddr
//    * WithListener
//
// See the documentation of the individual options for details. If the same
// option is passed more than once, the latest passed option wins. Unsupported
// options are silently ignored.
//
// If neither UseListener nor UseAddr are passed the server listens on
// 127.0.0.1:0 by default, i.e. it chooses a random port and listens on the
// loopback interface only.
//
// ListenAndServe blocks the current go routine.
func ListenAndServe(s Server, options ...Option) error {
	const op errors.Op = "netutil/ListenAndServe"

	nuOpts := applyOptions(&netutilOptions{}, options)

	addr := nuOpts.Addr
	if addr == "" {
		addr = "127.0.0.1:0"
	}

	l := nuOpts.Listener
	if l == nil {
		var err error

		if l, err = net.Listen("tcp", addr); err != nil {
			return errors.New(op, "listen", err)
		}
	}
	// We can't be sure all implementors of Server close l themselves. Thus
	// we close it as well.
	defer l.Close()

	if nuOpts.AddrChan != nil {
		go sendAddr(nuOpts.AddrChan, l)
	}

	if err := s.Serve(l); !isErrIgnored(err) {
		return errors.New(op, "serve", err)
	}
	return nil
}

func sendAddr(c chan<- string, l net.Listener) {
	c <- l.Addr().String()
}

func isErrIgnored(err error) bool {
	return err == nil || err == http.ErrServerClosed
}
