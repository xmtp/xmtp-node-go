package api

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	identityv1 "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	messagev1 "github.com/xmtp/xmtp-node-go/pkg/proto/message_api/v1"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/ratelimiter"
	"github.com/xmtp/xmtp-node-go/pkg/utils"
)

type RateLimitInterceptor struct {
	limiter     ratelimiter.RateLimiter
	ipAllowList authz.AllowList
	log         *zap.Logger
}

func NewRateLimitInterceptor(
	limiter ratelimiter.RateLimiter,
	ipAllowList authz.AllowList,
	log *zap.Logger,
) *RateLimitInterceptor {
	return &RateLimitInterceptor{limiter: limiter, ipAllowList: ipAllowList, log: log}
}

func (rli *RateLimitInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := rli.applyLimits(ctx, info.FullMethod, req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (rli *RateLimitInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		req interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := rli.applyLimits(stream.Context(), info.FullMethod, req); err != nil {
			return err
		}
		return handler(req, stream)
	}
}

func (rli *RateLimitInterceptor) applyLimits(
	ctx context.Context,
	fullMethod string,
	req interface{},
) error {
	// * for limit exhaustion return status.Errorf(codes.ResourceExhausted, ...)
	// * for other authorization failure return status.Errorf(codes.PermissionDenied, ...)
	_, method := splitMethodName(fullMethod)

	ip := utils.ClientIPFromContext(ctx)
	if len(ip) == 0 {
		// requests without an IP address are bucketed together as "ip_unknown"
		ip = "ip_unknown"
		rli.log.Warn("no ip found", logging.String("method", fullMethod))
	}

	permissions := rli.ipAllowList.GetPermission(ip)

	isPriority := permissions == authz.Priority
	isDenied := permissions == authz.Denied
	isAllowAll := permissions == authz.AllowAll

	if isDenied {
		// Debatable whether we want to return a message this clear or something more ambiguous
		return status.Error(codes.PermissionDenied, "request blocked")
	}

	if isAllowAll {
		return nil
	}

	cost := 1
	limitType := ratelimiter.DEFAULT
	switch req := req.(type) {
	case *mlsv1.RegisterInstallationRequest, *mlsv1.UploadKeyPackageRequest, *identityv1.PublishIdentityUpdateRequest:
		limitType = ratelimiter.PUBLISH
	case *mlsv1.SendWelcomeMessagesRequest:
		cost = len(req.Messages)
		limitType = ratelimiter.PUBLISH
	case *mlsv1.SendGroupMessagesRequest:
		cost = len(req.Messages)
		limitType = ratelimiter.PUBLISH
	case *messagev1.PublishRequest:
		cost = len(req.Envelopes)
		limitType = ratelimiter.PUBLISH
	case *messagev1.BatchQueryRequest:
		cost = len(req.Requests)
	}

	var bucket string
	if isPriority {
		bucket = "P" + ip + string(limitType)
	} else {
		bucket = "R" + ip + string(limitType)
	}

	if err := rli.limiter.Spend(limitType, bucket, uint16(cost), isPriority); err != nil {
		rli.log.Info("rate limited",
			logging.String("client_ip", ip),
			logging.Bool("priority", isPriority),
			logging.String("method", method),
			logging.String("limit", string(limitType)),
			logging.Int("cost", cost))

		return status.Error(codes.ResourceExhausted, err.Error())
	}

	return nil
}
