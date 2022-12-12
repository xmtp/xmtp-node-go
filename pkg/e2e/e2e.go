package e2e

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
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
	APIURL                  string
	DelayBetweenRunsSeconds int
	GitCommit               string
	MetricsPort             int
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
		s.newTest("messagev1 publish batch query", s.testMessageV1PublishBatchQuery),
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
