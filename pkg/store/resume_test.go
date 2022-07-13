package store

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/stretchr/testify/require"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
)

func TestStore_FindLastSeen(t *testing.T) {
	pubSubTopic := "test"

	msg1 := test.NewMessage("topic1", 1, "msg1")
	msg2 := test.NewMessage("topic2", 2, "msg2")
	msg3 := test.NewMessage("topic3", 3, "msg3")
	msg4 := test.NewMessage("topic4", 4, "msg4")
	msg5 := test.NewMessage("topic5", 5, "msg5")

	s, cleanup := newTestStore(t)
	defer cleanup()

	storeMessage(t, s, msg1, pubSubTopic)
	storeMessage(t, s, msg3, pubSubTopic)
	storeMessage(t, s, msg5, pubSubTopic)
	storeMessage(t, s, msg2, pubSubTopic)
	storeMessage(t, s, msg4, pubSubTopic)

	lastSeen, err := s.findLastSeen()
	require.NoError(t, err)
	require.Equal(t, msg5.Timestamp, lastSeen)
}

func TestStore_Resume_FromPeer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, cleanup := newTestStore(t)
	defer cleanup()

	pubSubTopic := "test"

	msgs := []*pb.WakuMessage{
		test.NewMessage("topic1", 1, "msg1"),
		test.NewMessage("topic1", 2, "msg2"),
		test.NewMessage("topic1", 3, "msg3"),
		test.NewMessage("topic1", 4, "msg4"),
		test.NewMessage("topic1", 5, "msg5"),
		test.NewMessage("topic2", 6, "msg6"),
		test.NewMessage("topic2", 7, "msg7"),
		test.NewMessage("topic2", 8, "msg8"),
		test.NewMessage("topic2", 9, "msg9"),
		test.NewMessage("topic2", 10, "msg10"),
	}

	for _, msg := range msgs {
		storeMessage(t, s1, msg, pubSubTopic)
	}

	s2, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s2.host, s1.host)

	expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{})

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{s1.host.ID()})
	require.NoError(t, err)
	require.Equal(t, 10, msgCount)

	expectMessages(t, s2, pubSubTopic, msgs)

	// Test duplication
	msgCount, err = s2.Resume(ctx, pubSubTopic, []peer.ID{s1.host.ID()})
	require.NoError(t, err)
	require.Equal(t, 0, msgCount)
}

func TestStore_Resume_WithListOfPeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubSubTopic := "test"
	invalidHost := test.NewPeer(t) // without store protocol

	s1, cleanup := newTestStore(t)
	defer cleanup()

	msg0 := &pb.WakuMessage{ContentTopic: "topic2"}
	storeMessage(t, s1, msg0, pubSubTopic)

	s2, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s2.host, s1.host)

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{invalidHost.ID(), s1.host.ID()})
	require.NoError(t, err)
	require.Equal(t, 1, msgCount)

	expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{msg0})
}

func TestStore_Resume_WithoutSpecifyingPeer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubSubTopic := "test"

	s1, cleanup := newTestStore(t)
	defer cleanup()

	msg0 := &pb.WakuMessage{ContentTopic: "topic2"}
	storeMessage(t, s1, msg0, pubSubTopic)

	s2, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s2.host, s1.host)

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{})
	require.NoError(t, err)
	require.Equal(t, 1, msgCount)

	expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{msg0})
}

func TestStore_Resume_MultiplePeersDifferentData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubSubTopic := "test"

	s1, cleanup := newTestStore(t)
	defer cleanup()

	s2, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s1.host, s2.host)

	s3, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s1.host, s3.host)

	msgsS2 := []*pb.WakuMessage{
		test.NewMessage("topic1", 1, "msg1"),
		test.NewMessage("topic1", 2, "msg2"),
		test.NewMessage("topic2", 3, "msg3"),
		test.NewMessage("topic2", 4, "msg4"),
		test.NewMessage("topic3", 5, "msg5"),
	}
	for _, msg := range msgsS2 {
		storeMessage(t, s2, msg, pubSubTopic)
	}

	msgsS3 := []*pb.WakuMessage{
		test.NewMessage("topic1", 1, "msg1"),
		test.NewMessage("topic1", 2, "msg2"),
		test.NewMessage("topic2", 3, "msg3"),
		test.NewMessage("topic3", 4, "msg4"),
		test.NewMessage("topic4", 6, "msg6"),
	}
	for _, msg := range msgsS3 {
		storeMessage(t, s3, msg, pubSubTopic)
	}

	msgCount, err := s1.Resume(ctx, pubSubTopic, []peer.ID{s2.host.ID(), s3.host.ID()})
	require.NoError(t, err)
	require.Equal(t, 7, msgCount)

	expectMessages(t, s1, pubSubTopic, []*pb.WakuMessage{
		test.NewMessage("topic1", 1, "msg1"),
		test.NewMessage("topic1", 2, "msg2"),
		test.NewMessage("topic2", 3, "msg3"),
		test.NewMessage("topic2", 4, "msg4"),
		test.NewMessage("topic3", 4, "msg4"),
		test.NewMessage("topic3", 5, "msg5"),
		test.NewMessage("topic4", 6, "msg6"),
	})
}

func TestStore_Resume_Paginated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, cleanup := newTestStore(t)
	defer cleanup()

	pubSubTopic := "test"

	msgs := []*pb.WakuMessage{
		test.NewMessage("topic1", 1, "msg1"),
		test.NewMessage("topic1", 2, "msg2"),
		test.NewMessage("topic1", 3, "msg3"),
		test.NewMessage("topic1", 4, "msg4"),
		test.NewMessage("topic1", 5, "msg5"),
		test.NewMessage("topic2", 6, "msg6"),
		test.NewMessage("topic2", 7, "msg7"),
		test.NewMessage("topic2", 8, "msg8"),
		test.NewMessage("topic2", 9, "msg9"),
		test.NewMessage("topic2", 10, "msg10"),
	}

	for _, msg := range msgs {
		storeMessage(t, s1, msg, pubSubTopic)
	}

	s2, cleanup := newTestStore(t)
	s2.resumePageSize = 2
	defer cleanup()
	addStoreProtocol(t, s2.host, s1.host)

	expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{})

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{s1.host.ID()})
	require.NoError(t, err)
	require.Equal(t, 10, msgCount)
	expectMessages(t, s2, pubSubTopic, msgs)
}

func TestStore_Resume_StartTime(t *testing.T) {
	offset := int64(100 * time.Second)
	tcs := []struct {
		name          string
		startTime     int64
		storedMsgs    []*pb.WakuMessage
		expectedMsgs  []*pb.WakuMessage
		expectedCount int
	}{
		{
			name:      "negative unset",
			startTime: -1,
			storedMsgs: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("topic1", 2, "msg2"),
				test.NewMessage("topic1", 3, "msg3"),
			},
			expectedMsgs: []*pb.WakuMessage{
				test.NewMessage("topic1", 1, "msg1"),
				test.NewMessage("topic1", 2, "msg2"),
				test.NewMessage("topic1", 3, "msg3"),
			},
			expectedCount: 3,
		},
		{
			name:      "beginning of time",
			startTime: 0,
			storedMsgs: []*pb.WakuMessage{
				test.NewMessage("topic1", offset+1, "msg1"),
				test.NewMessage("topic1", offset+2, "msg2"),
				test.NewMessage("topic1", offset+3, "msg3"),
			},
			expectedMsgs: []*pb.WakuMessage{
				test.NewMessage("topic1", offset+1, "msg1"),
				test.NewMessage("topic1", offset+2, "msg2"),
				test.NewMessage("topic1", offset+3, "msg3"),
			},
			expectedCount: 3,
		},
		{
			name:      "recent past",
			startTime: offset + 2,
			storedMsgs: []*pb.WakuMessage{
				test.NewMessage("topic1", offset+1, "msg1"),
				test.NewMessage("topic1", offset+2, "msg2"),
				test.NewMessage("topic1", offset+3, "msg3"),
			},
			expectedMsgs: []*pb.WakuMessage{
				test.NewMessage("topic1", offset+2, "msg2"),
				test.NewMessage("topic1", offset+3, "msg3"),
			},
			expectedCount: 2,
		},
		{
			name:      "future",
			startTime: offset + 4,
			storedMsgs: []*pb.WakuMessage{
				test.NewMessage("topic1", offset+1, "msg1"),
				test.NewMessage("topic1", offset+2, "msg2"),
				test.NewMessage("topic1", offset+3, "msg3"),
			},
			expectedMsgs:  []*pb.WakuMessage{},
			expectedCount: 0,
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pubSubTopic := "test-" + test.RandomStringLower(5)

			s1, cleanup := newTestStore(t, WithResumeStartTime(tc.startTime))
			defer cleanup()

			for _, msg := range tc.storedMsgs {
				storeMessage(t, s1, msg, pubSubTopic)
			}
			expectMessages(t, s1, pubSubTopic, tc.storedMsgs)

			s2, cleanup := newTestStore(t, WithResumeStartTime(tc.startTime))
			defer cleanup()
			addStoreProtocol(t, s2.host, s1.host)
			expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{})

			msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{s1.host.ID()})
			require.NoError(t, err)
			require.Equal(t, tc.expectedCount, msgCount)
			expectMessages(t, s2, pubSubTopic, tc.expectedMsgs)
		})
	}
}
