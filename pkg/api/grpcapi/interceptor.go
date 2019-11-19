package grpcapi

import (
	"context"

	"github.com/fhofherr/acmeproxy/pkg/api/grpcapi/internal/pb"
	"github.com/fhofherr/acmeproxy/pkg/errors"
	"github.com/fhofherr/golf/log"
	"google.golang.org/grpc"
)

type unaryServerInterceptor struct {
	TokenParser TokenParser
	Logger      log.Logger
}

func (u *unaryServerInterceptor) intercept(
	ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (
	res interface{}, err error,
) {
	ctx, err = tokenAuthCtx(ctx, u.TokenParser)
	if err != nil {
		errors.Log(u.Logger, err)
		err = pb.ToGRPCStatusError(err)
		return
	}
	errors.LogFunc(u.Logger, func() error {
		res, err = handler(ctx, req)
		return pb.FromGRPCStatusError(err)
	})

	return
}
