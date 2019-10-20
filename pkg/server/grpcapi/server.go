package grpcapi

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/acmeproxy/pkg/server/grpcapi/internal/pb"
	"google.golang.org/grpc"
)

// Server represents the grpc API andler.
type Server struct {
	grpcServer *grpc.Server
	once       sync.Once
	started    uint32
}

// Serve accepts incomming connections.
//
// Calling Serve blocks the current go routine
func (s *Server) Serve(l net.Listener) error {
	const op errors.Op = "grpcapi/server.Serve"

	s.initialize()
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New(op, "already started")
	}
	return errors.Wrap(s.grpcServer.Serve(l), op, "serve")
}

func (s *Server) initialize() {
	s.once.Do(func() {
		s.grpcServer = grpc.NewServer()
		pb.RegisterAdminServer(s.grpcServer, &adminServer{})
	})
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
