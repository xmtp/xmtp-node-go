package main

import (
	"context"
	"crypto/ecdsa"
	"os"
	"strconv"
	"strings"

	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/e2e"
	"go.uber.org/zap"
)

const (
	localNetworkEnv = "local"
	localNodesURL   = "http://localhost:8000"
	remoteNodesURL  = "http://nodes.xmtp.com"
)

var (
	remoteAPIURLByEnv = map[string]string{
		"local":      "http://localhost:8080",
		"dev":        "https://api.dev.xmtp.network",
		"production": "https://api.production.xmtp.network",
	}

	GitCommit string // set during build
)

func main() {
	ctx := context.Background()
	log, err := zap.NewProduction()
	requireNoError(err)

	networkEnv := envVar("XMTPD_E2E_ENV", localNetworkEnv)
	nodesURL := envVar("XMTPD_E2E_NODES_URL", localNodesURL)
	if networkEnv != localNetworkEnv {
		nodesURL = remoteNodesURL
	}

	apiURL := envVar("XMTPD_E2E_API_URL", remoteAPIURLByEnv[networkEnv])

	v2WalletKeyHex := envVar("XMTPD_E2E_V2_AUTH_WALLET_KEY", "")
	var v2WalletKey *ecdsa.PrivateKey
	if v2WalletKeyHex != "" {
		v2WalletKey, err = crypto.HexToECDSA(v2WalletKeyHex)
		requireNoError(err)
	}

	v2IdentityKeyHex := envVar("XMTPD_E2E_V2_AUTH_IDENTITY_KEY", "")
	var v2IdentityKey *ecdsa.PrivateKey
	if v2IdentityKeyHex != "" {
		v2IdentityKey, err = crypto.HexToECDSA(v2IdentityKeyHex)
		requireNoError(err)
	}

	runner := e2e.NewRunner(ctx, log, &e2e.Config{
		Continuous:              envVarBool("E2E_CONTINUOUS"),
		ContinuousExitOnError:   envVarBool("E2E_CONTINUOUS_EXIT_ON_ERROR"),
		NetworkEnv:              networkEnv,
		BootstrapAddrs:          envVarStrings("XMTPD_E2E_BOOTSTRAP_ADDRS"),
		NodesURL:                nodesURL,
		APIURL:                  apiURL,
		DelayBetweenRunsSeconds: envVarInt("XMTPD_E2E_DELAY", 5),
		GitCommit:               GitCommit,
		V2WalletKey:             v2WalletKey,
		V2InstallationKey:       v2IdentityKey,
	})

	err = runner.Start()
	if err != nil {
		log.Fatal("running e2e", zap.Error(err))
	}
}

func requireNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func envVar(name, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	return val
}

func envVarStrings(name string) []string {
	val := os.Getenv(name)
	vals := strings.Split(val, ",")
	retVals := make([]string, 0, len(vals))
	for _, v := range vals {
		if v == "" {
			continue
		}
		retVals = append(retVals, v)
	}
	return retVals
}

func envVarBool(name string) bool {
	valStr := os.Getenv(name)
	return valStr != ""
}

func envVarInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
