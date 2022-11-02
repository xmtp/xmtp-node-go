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
	appVersionMetadataKey    = "x-app-version"
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
	clientName, _, clientVersion := parseVersionHeaderValue(md.Get(clientVersionMetadataKey))
	appName, _, appVersion := parseVersionHeaderValue(md.Get(appVersionMetadataKey))
	ti.log.Info("api request",
		zap.String("service", serviceName),
		zap.String("method", methodName),
		zap.String("client", clientName),
		zap.String("client_version", clientVersion),
		zap.String("app", appName),
		zap.String("app_version", appVersion),
	)
	metrics.EmitAPIRequest(ctx, serviceName, methodName, clientName, clientVersion, appName, appVersion)
	return nil
}

func splitMethodName(fullMethodName string) (serviceName string, methodName string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func parseVersionHeaderValue(vals []string) (name string, version string, full string) {
	if len(vals) == 0 {
		return
	}
	full = vals[0]
	parts := strings.Split(full, "/")
	if len(parts) > 0 {
		name = parts[0]
		if len(parts) > 1 {
			version = parts[1]
		}
	}
	return
}
