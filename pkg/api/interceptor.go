package api

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	messagev1 "github.com/xmtp/proto/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"github.com/xmtp/xmtp-node-go/pkg/types"
)

var ErrDenyListed = errors.New("wallet is deny listed")

// WalletAuthorizer implements the authentication/authorization flow of client requests.
// It is intended to be hooked up with a GRPC server as an interceptor.
// It requires all requests to include an Authorization: Bearer header
// carrying a base-64 encoded messagev1.Token.
// The token ties the request to a wallet (authentication).
// Authorization decisions are then based on the authenticated wallet.
type WalletAuthorizer struct {
	*AuthnConfig
	privilegedAddresses map[types.WalletAddr]bool
}

// NewWalletAuthorizer creates an authorizer configured based on the Config.
func NewWalletAuthorizer(config *AuthnConfig) *WalletAuthorizer {
	privilegedAddresses := make(map[types.WalletAddr]bool)
	for _, address := range config.PrivilegedAddresses {
		privilegedAddresses[types.WalletAddr(address)] = true
	}
	return &WalletAuthorizer{AuthnConfig: config, privilegedAddresses: privilegedAddresses}
}

func (wa *WalletAuthorizer) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		if wa.requiresAuthorization(req) {
			if err := wa.authorize(ctx, req); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

func (wa *WalletAuthorizer) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// TODO(mk): Add metrics
		return handler(srv, stream)
	}
}

func (wa *WalletAuthorizer) requiresAuthorization(req interface{}) bool {
	_, isPublish := req.(*messagev1.PublishRequest)
	return isPublish
}

func (wa *WalletAuthorizer) authorize(ctx context.Context, req interface{}) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get(authorizationMetadataKey)
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	words := strings.SplitN(values[0], " ", 2)
	if len(words) != 2 {
		return status.Errorf(codes.Unauthenticated, "invalid authorization header")
	}
	if scheme := strings.TrimSpace(words[0]); scheme != "Bearer" {
		return status.Errorf(codes.Unauthenticated, "unrecognized authorization scheme %s", scheme)
	}
	token, err := decodeToken(strings.TrimSpace(words[1]))
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "extracting token: %s", err)
	}

	wallet, err := validateToken(ctx, wa.Log, token, time.Now())
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "validating token: %s", err)
	}

	if pub, isPublish := req.(*messagev1.PublishRequest); isPublish {
		for _, env := range pub.Envelopes {
			if !wa.privilegedAddresses[wallet] && !allowedToPublish(env.ContentTopic, wallet) {
				return status.Errorf(codes.PermissionDenied, "publishing to restricted topic")
			}
		}
	}

	return wa.authorizeWallet(ctx, wallet)
}

func (wa *WalletAuthorizer) authorizeWallet(ctx context.Context, wallet types.WalletAddr) error {
	// * for limit exhaustion return status.Errorf(codes.ResourceExhausted, ...)
	// * for other authorization failure return status.Errorf(codes.PermissionDenied, ...)

	var allowListed bool
	if wa.AllowLists {
		if wa.AllowLister.IsDenyListed(wallet.String()) {
			wa.Log.Debug("wallet deny listed", logging.WalletAddress(wallet.String()))
			return status.Errorf(codes.PermissionDenied, ErrDenyListed.Error())
		}
		allowListed = wa.AllowLister.IsAllowListed(wallet.String())
	}

	if !wa.Ratelimits {
		return nil
	}
	err := wa.Limiter.Spend(wallet.String(), allowListed)
	if err == nil {
		return nil
	}
	wa.Log.Debug("wallet rate limited", logging.WalletAddress(wallet.String()))
	return status.Errorf(codes.ResourceExhausted, err.Error())
}

func allowedToPublish(topic string, wallet types.WalletAddr) bool {
	parts := strings.Split(topic, "-")
	return len(parts) != 2 ||
		(parts[0] != "contact" && parts[0] != "privatestore") ||
		types.WalletAddr(parts[1]) == wallet
}
