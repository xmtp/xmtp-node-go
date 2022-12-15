package e2e

import (
	"context"
	"math/rand"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type Suite struct {
	ctx  context.Context
	log  *zap.Logger
	rand *rand.Rand

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
}

type testRunFunc func(log *zap.Logger) error

type Test struct {
	Name string
	Run  testRunFunc
}

func NewSuite(ctx context.Context, log *zap.Logger, config *Config) *Suite {
	e := &Suite{
		ctx:    ctx,
		log:    log,
		rand:   rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
		config: config,
	}
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
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[s.rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func withAuth(ctx context.Context) (context.Context, error) {
	token, _, err := api.GenerateToken(time.Now(), false)
	if err != nil {
		return ctx, err
	}
	et, err := api.EncodeToken(token)
	if err != nil {
		return ctx, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+et)
	return ctx, nil
}
