package e2e

import (
	"context"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	_ "net/http/pprof"
)

type Suite struct {
	ctx context.Context
	log *zap.Logger

	config *Config
}

type Config struct {
	Continuous              bool
	NetworkEnv              string
	BootstrapAddrs          []string
	NodesURL                string
	APIURL                  string
	DelayBetweenRunsSeconds int
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
		config: config,
	}
	return e
}

func (s *Suite) Tests() []*Test {
	return []*Test{
		s.newTest("messagev1 publish subscribe query", s.testMessageV1PublishSubscribeQuery),
		s.newTest("waku publish subscribe query", s.testWakuPublishSubscribeQuery),
	}
}

func (s *Suite) newTest(name string, runFn testRunFunc) *Test {
	return &Test{
		Name: name,
		Run:  runFn,
	}
}

func withAuth(ctx context.Context) (context.Context, error) {
	token, _, err := api.GenerateToken(time.Now())
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
