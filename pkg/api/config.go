package api

import (
	"net"
	"strconv"

	"github.com/pkg/errors"
	wakunode "github.com/waku-org/go-waku/waku/v2/node"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	mlsstore "github.com/xmtp/xmtp-node-go/pkg/mls/store"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	"github.com/xmtp/xmtp-node-go/pkg/ratelimiter"
	"github.com/xmtp/xmtp-node-go/pkg/store"
	"go.uber.org/zap"
)

var (
	ErrMissingLog   = errors.New("missing log config")
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
	EnableMls   bool         `long:"enable-mls" description:"Enable the MLS server"`
}

type Config struct {
	Options
	AllowLister  authz.WalletAllowLister
	Waku         *wakunode.WakuNode
	Log          *zap.Logger
	Store        *store.Store
	MLSStore     mlsstore.MlsStore
	MLSValidator mlsvalidate.MLSValidationService
}

// AuthnOptions bundle command line options associated with the authn package.
type AuthnOptions struct {
	/*
		Enable is the master switch for the authentication module.
		If it is false then the other options in this group are ignored.

		The module enforces authentication for requests that require it (currently Publish only).
		Authenticated requests will be permitted according to the rules of the request type,
		(i.e. you can't publish into other wallets' contact and private topics).
	*/
	Enable bool `long:"enable" description:"require client authentication via wallet tokens"`
	/*
		Ratelimits enables request rate limiting.

		Requests are bucketed by client IP address and request type (there is one bucket for all requests without IPs).
		Each bucket is allocated a number of tokens that are refilled at a fixed rate per minute
		up to a given maximum number of tokens.
		Requests cost 1 token by default, except Publish requests cost the number of Envelopes carried
		and BatchQuery requests cost the number of queries carried.
		The limits depend on request type, e.g. Publish requests get lower limits than other types of request.
		If Allowlists is also true then requests with Bearer tokens from wallets explicitly Allowed get priority,
		i.e. a predefined multiple the configured limit.
		Priority wallets get separate IP buckets from regular wallets.
	*/
	Ratelimits bool `long:"ratelimits" description:"apply rate limits per client IP address"`
	/*
		Allowlists enables wallet allow lists.

		All requests that require authentication (currently Publish only) will be rejected
		for wallets that are set as Denied in the allow list.
		Wallets that are explicitly Allowed will get priority rate limits if Ratelimits is true.
	*/
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

func (params *Config) validate() error {
	if params.Log == nil {
		return ErrMissingLog
	}
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
