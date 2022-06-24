package store

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/status-im/go-waku/tests"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/stretchr/testify/require"
)

func TestStore_FindLastSeenMessage(t *testing.T) {
	pubSubTopic := "test"

	msg1 := tests.CreateWakuMessage("topic1", 1)
	msg2 := tests.CreateWakuMessage("topic2", 2)
	msg3 := tests.CreateWakuMessage("topic3", 3)
	msg4 := tests.CreateWakuMessage("topic4", 4)
	msg5 := tests.CreateWakuMessage("topic5", 5)

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
		tests.CreateWakuMessage("topic1", 1),
		tests.CreateWakuMessage("topic1", 2),
		tests.CreateWakuMessage("topic1", 3),
		tests.CreateWakuMessage("topic1", 4),
		tests.CreateWakuMessage("topic1", 5),
		tests.CreateWakuMessage("topic2", 6),
		tests.CreateWakuMessage("topic2", 7),
		tests.CreateWakuMessage("topic2", 8),
		tests.CreateWakuMessage("topic2", 9),
		tests.CreateWakuMessage("topic2", 10),
	}

	for _, msg := range msgs {
		storeMessage(t, s1, msg, pubSubTopic)
	}

	s2, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s2.h, s1.h)

	expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{})

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{s1.h.ID()})
	require.NoError(t, err)
	require.Equal(t, 10, msgCount)

	expectMessages(t, s2, pubSubTopic, msgs)

	// Test duplication
	msgCount, err = s2.Resume(ctx, pubSubTopic, []peer.ID{s1.h.ID()})
	require.NoError(t, err)
	require.Equal(t, 0, msgCount)
}

func TestStore_Resume_WithListOfPeers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubSubTopic := "test"
	invalidHost := newTestPeer(t) // without store protocol

	s1, cleanup := newTestStore(t)
	defer cleanup()

	msg0 := &pb.WakuMessage{ContentTopic: "topic2"}
	storeMessage(t, s1, msg0, pubSubTopic)

	s2, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s2.h, s1.h)

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{invalidHost.ID(), s1.h.ID()})
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
	addStoreProtocol(t, s2.h, s1.h)

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
	addStoreProtocol(t, s1.h, s2.h)

	s3, cleanup := newTestStore(t)
	defer cleanup()
	addStoreProtocol(t, s1.h, s3.h)

	msgsS2 := []*pb.WakuMessage{
		tests.CreateWakuMessage("topic1", 1),
		tests.CreateWakuMessage("topic1", 2),
		tests.CreateWakuMessage("topic2", 3),
		tests.CreateWakuMessage("topic2", 4),
		tests.CreateWakuMessage("topic3", 5),
	}
	for _, msg := range msgsS2 {
		storeMessage(t, s2, msg, pubSubTopic)
	}

	msgsS3 := []*pb.WakuMessage{
		tests.CreateWakuMessage("topic1", 1),
		tests.CreateWakuMessage("topic1", 2),
		tests.CreateWakuMessage("topic2", 3),
		tests.CreateWakuMessage("topic3", 4),
		tests.CreateWakuMessage("topic4", 6),
	}
	for _, msg := range msgsS3 {
		storeMessage(t, s3, msg, pubSubTopic)
	}

	msgCount, err := s1.Resume(ctx, pubSubTopic, []peer.ID{s2.h.ID(), s3.h.ID()})
	require.NoError(t, err)
	require.Equal(t, 7, msgCount)

	expectMessages(t, s1, pubSubTopic, []*pb.WakuMessage{
		tests.CreateWakuMessage("topic1", 1),
		tests.CreateWakuMessage("topic1", 2),
		tests.CreateWakuMessage("topic2", 3),
		tests.CreateWakuMessage("topic2", 4),
		tests.CreateWakuMessage("topic3", 4),
		tests.CreateWakuMessage("topic3", 5),
		tests.CreateWakuMessage("topic4", 6),
	})
}

func TestStore_Resume_Paginated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, cleanup := newTestStore(t)
	defer cleanup()

	pubSubTopic := "test"

	msgs := []*pb.WakuMessage{
		tests.CreateWakuMessage("topic1", 1),
		tests.CreateWakuMessage("topic1", 2),
		tests.CreateWakuMessage("topic1", 3),
		tests.CreateWakuMessage("topic1", 4),
		tests.CreateWakuMessage("topic1", 5),
		tests.CreateWakuMessage("topic2", 6),
		tests.CreateWakuMessage("topic2", 7),
		tests.CreateWakuMessage("topic2", 8),
		tests.CreateWakuMessage("topic2", 9),
		tests.CreateWakuMessage("topic2", 10),
	}

	for _, msg := range msgs {
		storeMessage(t, s1, msg, pubSubTopic)
	}

	s2, cleanup := newTestStore(t)
	s2.resumePageSize = 2
	defer cleanup()
	addStoreProtocol(t, s2.h, s1.h)

	expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{})

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{s1.h.ID()})
	require.NoError(t, err)
	require.Equal(t, 10, msgCount)
	expectMessages(t, s2, pubSubTopic, msgs)
}
