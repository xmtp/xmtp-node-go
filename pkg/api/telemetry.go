package api

import (
	"context"
	"strings"

	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	clientVersionMetadataKey = "x-client-version"
)

type TelemetryInterceptor struct {
	log *zap.Logger
}

func NewTelemetryInterceptor(log *zap.Logger) *TelemetryInterceptor {
	return &TelemetryInterceptor{
		log: log,
	}
}

func (ti *TelemetryInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		err := ti.execute(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (ti *TelemetryInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := ti.execute(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

func (ti *TelemetryInterceptor) execute(ctx context.Context, fullMethod string) error {
	serviceName, methodName := splitMethodName(fullMethod)
	md, _ := metadata.FromIncomingContext(ctx)
	vals := md.Get(clientVersionMetadataKey)
	var clientVersion string
	if len(vals) > 0 {
		clientVersion = vals[0]
	}
	ti.log.Info("api request",
		zap.String("service", serviceName),
		zap.String("method", methodName),
		zap.String("client_version", clientVersion),
	)
	metrics.EmitAPIRequest(ctx, serviceName, methodName, clientVersion)
	return nil
}

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}
