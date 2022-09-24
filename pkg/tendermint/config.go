package tendermint

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"
	"go.uber.org/zap"
)

type Options struct {
	ConfigPath           string `long:"config-path" description:"Tendermint config file path" default:".tendermint/config/config.toml"`
	DBPath               string `long:"db-path" description:"Tendermint DB path" default:".tendermint/badger"`
	P2PListenAddress     string `long:"p2p-listen-address" description:"Tendermint P2P listen address" default:"tcp://0.0.0.0:26656"`
	RPCListenAddress     string `long:"rpc-listen-address" description:"Tendermint RPC listen address" default:"tcp://0.0.0.0:26657"`
	RPCGRPCListenAddress string `long:"rpc-grpc-listen-address" description:"Tendermint RPC gRPC listen address" default:"tcp://0.0.0.0:26658"`
	MempoolRootDir       string `long:"mempool-root-dir" description:"Tendermint mempool root dir" default:".tendermint/data/mempool"`
	ConsensusRootDir     string `long:"consensus-root-dir" description:"Tendermint consensus root dir" default:".tendermint/data/consensus"`
	ProxyAppAddress      string `long:"proxy-app-address" description:"Tendermint proxy app listen address" default:"tcp://127.0.0.1:26658"`
}

type Config struct {
	Options
	Log        *zap.Logger
	BaseConfig *tmconfig.Config
}

func (c *Config) TendermintConfig() (*tmconfig.Config, error) {
	cfg := c.BaseConfig
	if cfg == nil {
		cfg = tmconfig.DefaultConfig()
	}

	cfg.RootDir = filepath.Dir(filepath.Dir(c.ConfigPath))
	viper := viper.New()
	viper.SetConfigFile(c.ConfigPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper failed to read config file: %w", err)
	}
	if err := viper.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("viper failed to unmarshal config: %w", err)
	}
	if err := cfg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}

	cfg.ProxyApp = c.ProxyAppAddress

	cfg.RPC.GRPCListenAddress = fmt.Sprintf(c.RPCGRPCListenAddress)
	cfg.RPC.ListenAddress = fmt.Sprintf(c.RPCListenAddress)
	cfg.P2P.ListenAddress = fmt.Sprintf(c.P2PListenAddress)

	cfg.Mempool.CacheSize = 0

	cfg.Mempool.RootDir = c.MempoolRootDir
	err := os.MkdirAll(cfg.Mempool.RootDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	cfg.Consensus.RootDir = c.ConsensusRootDir
	err = os.MkdirAll(cfg.Consensus.RootDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
