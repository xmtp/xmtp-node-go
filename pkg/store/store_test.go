package store

import (
	"testing"

	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

type testStoreOption func(c *Config)

func WithCleaner(opts CleanerOptions) testStoreOption {
	return func(c *Config) {
		c.Options.Cleaner = opts
	}
}

func newTestStore(t *testing.T, opts ...testStoreOption) (*Store, func()) {
	log := test.NewLog(t)
	db, _, dbCleanup := test.NewDB(t)

	c := &Config{
		Log:       log,
		DB:        db,
		ReaderDB:  db,
		CleanerDB: db,
	}
	for _, opt := range opts {
		opt(c)
	}

	store, err := New(c)
	require.NoError(t, err)

	return store, func() {
		store.Close()
		dbCleanup()
	}
}

func storeMessage(t *testing.T, s *Store, env *messagev1.Envelope) {
	_, err := s.InsertMessage(env)
	require.NoError(t, err)
}

func storeMessageWithTime(t *testing.T, s *Store, env *messagev1.Envelope, receiverTime int64) {
	err := s.insertMessage(env, receiverTime)
	require.NoError(t, err)
}
