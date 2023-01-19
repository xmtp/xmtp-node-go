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

	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
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

const (
	TopicCategoryPermissionUnknown        = 0
	TopicCategoryPermissionAuthRequired   = 1
	TopicCategoryPermissionNoAuthRequired = 2
)

var TopicCategoryPermissions = map[string]int{
	"privatestore": TopicCategoryPermissionAuthRequired,
	"contact":      TopicCategoryPermissionAuthRequired,
	"m":            TopicCategoryPermissionNoAuthRequired,
	"dm":           TopicCategoryPermissionNoAuthRequired,
	"intro":        TopicCategoryPermissionNoAuthRequired,
	"invite":       TopicCategoryPermissionNoAuthRequired,
}

func allowedToPublish(topic string, wallet types.WalletAddr) bool {
	// The `topic` contains something like
	//   "/xmtp/0/privatestore-0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266/key_bundle/proto"
	//   "/xmtp/0/contact-0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266/proto"
	//   "/xmtp/0/m-111122223333444455556666/proto"

	// The rules only apply to things that start with "/xmtp/0/..."
	if !strings.HasPrefix(topic, "/xmtp/0/") {
		return true
	}
	// e.g. segment[3] = "privatestore-0x12345.." or "m-11-22-33..."
	segments := strings.Split(topic, "/")

	// e.g. parts[0] = "privatestore" or "dm"
	// note: some segments have multiple parts ("m-11-22-33...")
	parts := strings.Split(segments[3], "-")

	permission := TopicCategoryPermissions[parts[0]]

	if permission == TopicCategoryPermissionAuthRequired {
		// If auth is required, then the second part (after the hyphen)
		// must be the verified wallet address of the caller.
		if len(parts) < 2 || types.WalletAddr(parts[1]) != wallet {
			return false
		}
	}

	return true
}
