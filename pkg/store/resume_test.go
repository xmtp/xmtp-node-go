package store

import (
	"context"
	"testing"

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
	invalidHost := test.NewPeer(t) // without store protocol

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

	msgCount, err := s1.Resume(ctx, pubSubTopic, []peer.ID{s2.h.ID(), s3.h.ID()})
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
	addStoreProtocol(t, s2.h, s1.h)

	expectMessages(t, s2, pubSubTopic, []*pb.WakuMessage{})

	msgCount, err := s2.Resume(ctx, pubSubTopic, []peer.ID{s1.h.ID()})
	require.NoError(t, err)
	require.Equal(t, 10, msgCount)
	expectMessages(t, s2, pubSubTopic, msgs)
}
