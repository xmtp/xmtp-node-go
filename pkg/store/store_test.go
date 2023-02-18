package store

import (
	"testing"

	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/utils"
	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func newTestStore(t *testing.T, opts ...Option) (*XmtpStore, func()) {
	db, _, dbCleanup := test.NewDB(t)
	log := utils.Logger()

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

func storeMessage(t *testing.T, s *XmtpStore, msg *pb.WakuMessage, pubSubTopic string) {
	_, err := s.storeMessage(test.NewEnvelope(t, msg, pubSubTopic))
	require.NoError(t, err)
}

func storeMessageWithTime(t *testing.T, s *XmtpStore, msg *pb.WakuMessage, pubSubTopic string, receiverTime int64) {
	env := protocol.NewEnvelope(msg, receiverTime, pubSubTopic)
	_, err := s.storeMessage(env)
	require.NoError(t, err)
}
