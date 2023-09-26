package api

import (
	"context"
	"strings"

	apicontext "github.com/xmtp/xmtp-node-go/pkg/api/message/v1/context"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		err := handler(srv, stream)
		ti.record(stream.Context(), info.FullMethod, err)
		return err
	}
}

func (ti *TelemetryInterceptor) record(ctx context.Context, fullMethod string, err error) {
	serviceName, methodName := splitMethodName(fullMethod)
	ri := apicontext.NewRequesterInfo(ctx)
	fields := append(
		[]zapcore.Field{
			zap.String("service", serviceName),
			zap.String("method", methodName),
		},
		ri.ZapFields()...,
	)

	if ip := clientIPFromContext(ctx); len(ip) > 0 {
		fields = append(fields, zap.String("client_ip", ip))
	}

	logFn := ti.log.Debug

	if err != nil {
		logFn = ti.log.Info
		fields = append(fields, zap.Error(err))
		grpcErr, _ := status.FromError(err)
		if grpcErr != nil {
			errCode := grpcErr.Code().String()
			fields = append(fields, []zapcore.Field{
				zap.String("error_code", errCode),
				zap.String("error_message", grpcErr.Message()),
			}...)
		} else {
			fields = append(fields, []zapcore.Field{
				zap.String("error_code", codes.Internal.String()),
				zap.String("error_message", err.Error()),
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
