package authn

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	messagev1 "github.com/xmtp/proto/go/message_api/v1"
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
	*Config
}

// NewWalletAuthorizer creates an authorizer configured based on the Config.
func NewWalletAuthorizer(config *Config) *WalletAuthorizer {
	return &WalletAuthorizer{Config: config}
}

func (wa *WalletAuthorizer) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := wa.authorize(ctx); err != nil {
			return nil, err
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
		if err := wa.authorize(stream.Context()); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

func (wa *WalletAuthorizer) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md.Get("authorization")
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
	token, err := DecodeToken(strings.TrimSpace(words[1]))
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "extracting token: %s", err)
	}

	wallet, err := validateToken(ctx, token, time.Now())
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "validating token: %s", err)
	}

	return wa.authorizeWallet(ctx, wallet)
}

func (wa *WalletAuthorizer) authorizeWallet(ctx context.Context, wallet types.WalletAddr) error {
	// * for limit exhaustion return status.Errorf(codes.ResourceExhausted, ...)
	// * for other authorization failure return status.Errorf(codes.PermissionDenied, ...)
	if !wa.Ratelimits {
		return nil
	}
	var allowListed bool
	if wa.AllowLists {
		if wa.AllowLister.IsDenyListed(wallet.String()) {
			return status.Errorf(codes.PermissionDenied, ErrDenyListed.Error())
		}
		allowListed = wa.AllowLister.IsAllowListed(wallet.String())
	}
	err := wa.Limiter.Spend(wallet.String(), allowListed)
	if err == nil {
		return nil
	}
	return status.Errorf(codes.ResourceExhausted, err.Error())
}

// DecodeToken is exported for testing purposes.
// Not meant to be part of the package API.
func DecodeToken(s string) (*messagev1.Token, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var token messagev1.Token
	err = proto.Unmarshal(b, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// EncodeToken is exported for testing purposes.
// Not meant to be part of the package API.
func EncodeToken(token *messagev1.Token) (string, error) {
	b, err := proto.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
