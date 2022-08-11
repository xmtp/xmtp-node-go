package api

import (
	"net"
	"strconv"

	"github.com/pkg/errors"
	wakunode "github.com/status-im/go-waku/waku/v2/node"
	"github.com/xmtp/xmtp-node-go/pkg/authn"
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"go.uber.org/zap"
)

var (
	ErrMissingLog  = errors.New("missing log config")
	ErrMissingWaku = errors.New("missing waku config")
)

type Options struct {
	GRPCAddress string        `long:"grpc-address" description:"API GRPC listening address" default:"0.0.0.0"`
	GRPCPort    uint          `long:"grpc-port" description:"API GRPC listening port" default:"5556"`
	HTTPAddress string        `long:"http-address" description:"API HTTP listening address" default:"0.0.0.0"`
	HTTPPort    uint          `long:"http-port" description:"API HTTP listening port" default:"5555"`
	Authn       authn.Options `group:"API Authentication Options" namespace:"authn"`
}

type Config struct {
	Options
	Authorizer authz.WalletAllowLister
	Waku       *wakunode.WakuNode
	Log        *zap.Logger
}

func (params *Config) check() error {
	if params.Log == nil {
		return ErrMissingLog
	}
	if params.Waku == nil {
		return ErrMissingWaku
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
