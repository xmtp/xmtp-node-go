package api

import (
	"net"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/ratelimiter"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"go.uber.org/zap"
)

var (
	ErrMissingLog   = errors.New("missing log config")
	ErrMissingNATS  = errors.New("missing nats config")
	ErrMissingWaku  = errors.New("missing waku config")
	ErrMissingStore = errors.New("missing store config")
)

type Options struct {
	GRPCAddress string       `long:"grpc-address" description:"API GRPC listening address" default:"0.0.0.0"`
	GRPCPort    uint         `long:"grpc-port" description:"API GRPC listening port" default:"5556"`
	HTTPAddress string       `long:"http-address" description:"API HTTP listening address" default:"0.0.0.0"`
	HTTPPort    uint         `long:"http-port" description:"API HTTP listening port" default:"5555"`
	Authn       AuthnOptions `group:"API Authentication Options" namespace:"authn"`
	MaxMsgSize  int          `long:"max-msg-size" description:"Max message size in bytes (default 50MB)" default:"52428800"`
}

type Config struct {
	Options
	AllowLister authz.WalletAllowLister
	NATS        *nats.Conn
	Waku        *wakunode.WakuNode
	Log         *zap.Logger
	Store       *store.XmtpStore
}

// Options bundle command line options associated with the authn package.
type AuthnOptions struct {
	Enable              bool     `long:"enable" description:"require client authentication via wallet tokens"`
	Ratelimits          bool     `long:"ratelimits" description:"apply rate limits per wallet"`
	AllowLists          bool     `long:"allowlists" description:"apply higher limits for allow listed wallets (requires authz and ratelimits)"`
	PrivilegedAddresses []string `long:"privileged-address" description:"allow this address to publish into other user's topics"`
}

// Config bundles Options and other parameters needed to set up an authorizer.
type AuthnConfig struct {
	AuthnOptions
	Limiter     ratelimiter.RateLimiter
	AllowLister authz.WalletAllowLister
	Log         *zap.Logger
}

func (params *Config) check() error {
	if params.Log == nil {
		return ErrMissingLog
	}
	// if params.NATS == nil {
	// 	return ErrMissingNATS
	// }
	if params.Waku == nil {
		return ErrMissingWaku
	}
	if params.Store == nil {
		return ErrMissingStore
	}
	if err := validateAddr(params.HTTPAddress, params.HTTPPort); err != nil {
		return errors.Wrap(err, "Invalid HTTP Address")
	}
	if err := validateAddr(params.GRPCAddress, params.GRPCPort); err != nil {
		return errors.Wrap(err, "Invalid GRPC Address")
	}
	return nil
}

func validateAddr(addr string, port uint) error {
	_, err := net.ResolveTCPAddr("tcp", addrString(addr, port))
	return err
}

func addrString(addr string, port uint) string {
	return net.JoinHostPort(addr, strconv.Itoa(int(port)))
}
