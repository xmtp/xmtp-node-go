package api

import (
	"context"

	apicontext "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const IS_GATING_ENABLED = false

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
		ri := apicontext.NewRequesterInfo(ctx)
		if IS_GATING_ENABLED && !ri.IsSupportedClient {
			return nil, status.Errorf(codes.Unimplemented, "unsupported libxmtp version "+ri.LibxmtpVersion)
		}
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
		ri := apicontext.NewRequesterInfo(stream.Context())
		if IS_GATING_ENABLED && !ri.IsSupportedClient {
			return status.Errorf(codes.Unimplemented, "unsupported libxmtp version "+ri.LibxmtpVersion)
		}
		err := handler(srv, stream)
		return err
	}
}
