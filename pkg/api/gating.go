package api

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GatingInterceptor struct {
	log *zap.Logger
}

func NewGatingInterceptor(log *zap.Logger) *GatingInterceptor {
	return &GatingInterceptor{
		log: log,
	}
}

func (ti *GatingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		res, err := handler(ctx, req)
		return res, err
	}
}

func (ti *GatingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := handler(srv, stream)
		return err
	}
}
