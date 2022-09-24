package testing

import (
	"fmt"
	"path"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
	tmconfig "github.com/tendermint/tendermint/config"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	xtm "github.com/xmtp/xmtp-node-go/pkg/tendermint"
)

func NewTendermintNode(t *testing.T) (*xtm.Node, func()) {
	n, err := xtm.NewNode(&xtm.Config{
		Log:        NewLog(t),
		Options:    NewTendermintOptions(t),
		BaseConfig: tmconfig.TestConfig(),
	})
	require.NoError(t, err)
	require.NotNil(t, n)
	return n, func() {
		n.Close()
	}
}

func NewTendermintOptions(t *testing.T) xtm.Options {
	tmpDir := path.Join(t.TempDir(), RandomStringLower(12))
	tmRootDir := path.Join(tmpDir, ".tendermint")
	tmConfigPath := path.Join(tmRootDir, "config", "config.toml")
	tmDBPath := path.Join(tmpDir, ".badger")
	createTendermintConfig(t, tmRootDir)
	tmP2PPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	tmRPCPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	tmRPCGRPCPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	tmProxyAppPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	return xtm.Options{
		ConfigPath:           tmConfigPath,
		DBPath:               tmDBPath,
		ProxyAppAddress:      fmt.Sprintf("tcp://127.0.0.1:%d", tmProxyAppPort),
		P2PListenAddress:     fmt.Sprintf("tcp://127.0.0.1:%d", tmP2PPort),
		RPCListenAddress:     fmt.Sprintf("tcp://127.0.0.1:%d", tmRPCPort),
		RPCGRPCListenAddress: fmt.Sprintf("tcp://127.0.0.1:%d", tmRPCGRPCPort),
		MempoolRootDir:       path.Join(tmpDir, "data/mempool"),
		ConsensusRootDir:     path.Join(tmpDir, "data/consensus"),
	}
}

func createTendermintConfig(t *testing.T, rootDir string) {
	tmconfig.EnsureRoot(rootDir)

	cfg := tmconfig.TestConfig()

	cfg.RootDir = rootDir
	err := cfg.ValidateBasic()
	require.NoError(t, err)

	privValKeyFile := cfg.PrivValidatorKeyFile()
	privValStateFile := cfg.PrivValidatorStateFile()
	pv := privval.GenFilePV(privValKeyFile, privValStateFile)
	pv.Save()

	nodeKeyFile := cfg.NodeKeyFile()
	_, err = p2p.LoadOrGenNodeKey(nodeKeyFile)
	require.NoError(t, err)

	genFile := cfg.GenesisFile()
	genDoc := types.GenesisDoc{
		ChainID:         fmt.Sprintf("test-chain-%v", tmrand.Str(6)),
		GenesisTime:     tmtime.Now(),
		ConsensusParams: types.DefaultConsensusParams(),
	}
	pubKey, err := pv.GetPubKey()
	require.NoError(t, err)
	genDoc.Validators = []types.GenesisValidator{{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		Power:   10,
	}}

	err = genDoc.SaveAs(genFile)
	require.NoError(t, err)
}
