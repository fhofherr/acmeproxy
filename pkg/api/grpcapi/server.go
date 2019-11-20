package grpcapi

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/golf/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server represents the grpc API andler.
type Server struct {
	TokenParser TokenParser
	TLSConfig   *tls.Config
	Logger      log.Logger
	grpcServer  *grpc.Server
	once        sync.Once
	started     uint32
	initErr     error
}

// Serve accepts incomming connections.
//
// Calling Serve blocks the current go routine
func (s *Server) Serve(l net.Listener) error {
	const op errors.Op = "grpcapi/server.Serve"

	if err := s.initialize(); err != nil {
		return errors.New(op, err)
	}
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New(op, "already started")
	}
	return errors.Wrap(s.grpcServer.Serve(l), op, "serve")
}

func (s *Server) initialize() error {
	const op errors.Op = "grpcapi/server.initialize"

	s.once.Do(func() {
		if s.TokenParser == nil {
			s.initErr = errors.New(op, "no token parser provided")
			return
		}
		if s.TLSConfig == nil {
			s.initErr = errors.New(op, "no tls config provided")
			return
		}
		unaryInterceptor := &unaryServerInterceptor{
			TokenParser: s.TokenParser,
			Logger:      s.Logger,
		}
		creds := credentials.NewTLS(s.TLSConfig)
		s.grpcServer = grpc.NewServer(
			grpc.Creds(creds),
			grpc.UnaryInterceptor(unaryInterceptor.intercept),
		)
		pb.RegisterAdminServer(s.grpcServer, &adminServer{})
	})

	return s.initErr
}

// Shutdown performs a graceful shutdown of the Server.
func (s *Server) Shutdown(ctx context.Context) error {
	const op errors.Op = "grpcapi/server.Shutdown"

	if atomic.LoadUint32(&s.started) == 0 {
		return errors.New(op, "not started")
	}
	return errors.Wrap(s.stopGrpcServer(ctx), op)
}

func (s *Server) stopGrpcServer(ctx context.Context) error {
	const op errors.Op = "grpcapi/Server.stopGrpcServer"

	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.grpcServer.Stop()
		return errors.New(op, ctx.Err())
	}
}
