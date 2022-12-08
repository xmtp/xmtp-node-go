package api

import (
	"context"
	"strings"

	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
		res, err := handler(ctx, req)
		ti.record(ctx, info.FullMethod, err)
		return res, err
	}
}

func (ti *TelemetryInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		res := handler(srv, stream)
		ti.record(stream.Context(), info.FullMethod, nil)
		return res
	}
}

func (ti *TelemetryInterceptor) record(ctx context.Context, fullMethod string, err error) {
	serviceName, methodName := splitMethodName(fullMethod)
	md, _ := metadata.FromIncomingContext(ctx)
	clientName, _, clientVersion := parseVersionHeaderValue(md.Get(clientVersionMetadataKey))
	appName, _, appVersion := parseVersionHeaderValue(md.Get(appVersionMetadataKey))
	fields := []zapcore.Field{
		zap.String("service", serviceName),
		zap.String("method", methodName),
		zap.String("client", clientName),
		zap.String("client_version", clientVersion),
		zap.String("app", appName),
		zap.String("app_version", appVersion),
	}

	if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
		// There are potentially multiple comma separated IPs bundled in that first value
		ips := strings.Split(ips[0], ",")
		fields = append(fields, zap.String("client_ip", strings.TrimSpace(ips[0])))
	}

	logFn := ti.log.Info

	if err != nil {
		logFn = ti.log.Error
		fields = append(fields, zap.Error(err))
		grpcErr, _ := status.FromError(err)
		if grpcErr != nil {
			errCode := grpcErr.Code().String()
			fields = append(fields, []zapcore.Field{
				zap.String("error_code", errCode),
				zap.String("error_message", grpcErr.Message()),
			}...)
		}
	}

	logFn("api request", fields...)
	metrics.EmitAPIRequest(ctx, ti.log, fields)
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
