package e2e

import (
	"context"
	"crypto/ecdsa"
	"math/rand"
	"sync"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/xmtp/xmtp-node-go/pkg/api"
	"github.com/xmtp/xmtp-node-go/pkg/crypto"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type Suite struct {
	ctx    context.Context
	log    *zap.Logger
	rand   *rand.Rand
	randMu sync.Mutex

	config *Config
}

type Config struct {
	Continuous              bool
	ContinuousExitOnError   bool
	NetworkEnv              string
	BootstrapAddrs          []string
	NodesURL                string
	APIURL                  string
	DelayBetweenRunsSeconds int
	GitCommit               string

	V2WalletKey       *ecdsa.PrivateKey
	V2InstallationKey *ecdsa.PrivateKey
}

type testRunFunc func(log *zap.Logger) error

type Test struct {
	Name string
	Run  testRunFunc
}

func NewSuite(ctx context.Context, log *zap.Logger, config *Config) *Suite {
	if config.V2WalletKey == nil {
		log.Info("no v2 auth wallet key provided, generating a new one to use")
		config.V2WalletKey, _ = ethcrypto.GenerateKey()
	}
	v2WalletPublicKey := crypto.PublicKey(ethcrypto.FromECDSAPub(&config.V2WalletKey.PublicKey))

	if config.V2InstallationKey == nil {
		log.Info("no v2 auth identity key provided, generating a new one to use")
		config.V2InstallationKey, _ = ethcrypto.GenerateKey()
	}
	v2IdentityPublicKey := crypto.PublicKey(ethcrypto.FromECDSAPub(&config.V2InstallationKey.PublicKey))

	e := &Suite{
		ctx:    ctx,
		log:    log,
		rand:   rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
		config: config,
	}

	log.Info("e2e suite initialized",
		zap.String("v2_wallet_key_public_address", crypto.PublicKeyToAddress(v2WalletPublicKey).String()),
		zap.String("v2_identity_key_public_address", crypto.PublicKeyToAddress(v2IdentityPublicKey).String()),
	)
	return e
}

func (s *Suite) Tests() []*Test {
	return []*Test{
		s.newTest("messagev1 publish subscribe query", s.testMessageV1PublishSubscribeQuery),
	}
}

func (s *Suite) newTest(name string, runFn testRunFunc) *Test {
	return &Test{
		Name: name,
		Run:  runFn,
	}
}

func (s *Suite) randomStringLower(n int) string {
	s.randMu.Lock()
	defer s.randMu.Unlock()
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[s.rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func (s *Suite) withAuth(ctx context.Context) (context.Context, error) {
	token, _, err := api.BuildV2AuthToken(s.config.V2WalletKey, s.config.V2InstallationKey, time.Now())
	if err != nil {
		return ctx, err
	}
	et, err := api.EncodeAuthToken(token)
	if err != nil {
		return ctx, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+et)
	return ctx, nil
}
