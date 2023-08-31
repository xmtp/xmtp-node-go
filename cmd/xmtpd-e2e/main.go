package main

import (
	"context"
	"os"
	"strconv"

	"github.com/xmtp/xmtp-node-go/pkg/e2e"
	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"go.uber.org/zap"
)

const (
	localNetworkEnv = "local"
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
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	ctx := logging.With(context.Background(), log)

	networkEnv := envVar("XMTPD_E2E_ENV", localNetworkEnv)

	apiURL := envVar("XMTPD_E2E_API_URL", remoteAPIURLByEnv[networkEnv])

	runner := e2e.NewRunner(ctx, log, &e2e.Config{
		Continuous:              envVarBool("E2E_CONTINUOUS"),
		ContinuousExitOnError:   envVarBool("E2E_CONTINUOUS_EXIT_ON_ERROR"),
		NetworkEnv:              networkEnv,
		APIURL:                  apiURL,
		DelayBetweenRunsSeconds: envVarInt("XMTPD_E2E_DELAY", 5),
		GitCommit:               GitCommit,
	})

	err = runner.Start()
	if err != nil {
		log.Fatal("running e2e", zap.Error(err))
	}
}

func envVar(name, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	return val
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
