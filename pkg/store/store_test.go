package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func newTestStore(t *testing.T, opts ...Option) (*Store, func()) {
	db, _, dbCleanup := test.NewDB(t)
	log := test.NewLog(t)

	store, err := New(
		append(
			[]Option{
				WithLog(log),
				WithDB(db),
				WithReaderDB(db),
				WithCleanerDB(db),
			},
			opts...,
		)...,
	)
	require.NoError(t, err)

	return store, func() {
		store.Close()
		dbCleanup()
	}
}

func storeMessage(t *testing.T, s *Store, env *messagev1.Envelope) {
	_, err := s.storeMessage(env, time.Now().UTC().UnixNano())
	require.NoError(t, err)
}

func storeMessageWithTime(t *testing.T, s *Store, env *messagev1.Envelope, receiverTime int64) {
	_, err := s.storeMessage(env, receiverTime)
	require.NoError(t, err)
}
