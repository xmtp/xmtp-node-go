package filter

import (
	"context"
	"crypto/rand"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/status-im/go-waku/tests"
	v2 "github.com/status-im/go-waku/waku/v2"
	goWakuFilter "github.com/status-im/go-waku/waku/v2/protocol/filter"
	"github.com/status-im/go-waku/waku/v2/protocol/relay"
	"github.com/stretchr/testify/require"
)

func makeWakuRelay(t *testing.T, topic string, broadcaster v2.Broadcaster) (*relay.WakuRelay, *relay.Subscription, host.Host) {
	port, err := tests.FindFreePort(t, "", 5)
	require.NoError(t, err)

	host, err := tests.MakeHost(context.Background(), port, rand.Reader)
	require.NoError(t, err)

	relay, err := relay.NewWakuRelay(context.Background(), host, broadcaster, 0, tests.Logger())
	require.NoError(t, err)

	sub, err := relay.SubscribeToTopic(context.Background(), topic)
	require.NoError(t, err)

	return relay, sub, host
}

func makeWakuFilter(t *testing.T) (*WakuFilter, host.Host) {
	port, err := tests.FindFreePort(t, "", 5)
	require.NoError(t, err)

	host, err := tests.MakeHost(context.Background(), port, rand.Reader)
	require.NoError(t, err)

	filter, _ := NewWakuFilter(context.Background(), host, false, tests.Logger())

	return filter, host
}

func setupNodes(t *testing.T, topic string, ctx context.Context) (*WakuFilter, *WakuFilter, *relay.WakuRelay) {
	filter1, host1 := makeWakuFilter(t)
	broadcaster := v2.NewBroadcaster(10)
	relay, _, host2 := makeWakuRelay(t, topic, broadcaster)
	filter2, _ := NewWakuFilter(ctx, host2, true, tests.Logger())
	broadcaster.Register(filter2.MsgC)

	host1.Peerstore().AddAddr(host2.ID(), tests.GetHostAddress(host2), peerstore.PermanentAddrTTL)
	err := host1.Peerstore().AddProtocols(host2.ID(), string(FilterID_v10beta1))
	require.NoError(t, err)

	return filter1, filter2, relay
}

func TestWakuFilter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Test can't exceed 10 seconds
	defer cancel()

	testTopic := "/waku/2/go/filter/test"
	testContentTopic := "TopicA"

	filter1, filter2, node2Relay := setupNodes(t, testTopic, ctx)

	contentFilter := &goWakuFilter.ContentFilter{
		Topic:         string(testTopic),
		ContentTopics: []string{testContentTopic},
	}

	_, f, err := filter1.Subscribe(ctx, *contentFilter, goWakuFilter.WithPeer(filter2.h.ID()))
	require.NoError(t, err)

	// Sleep to make sure the filter is subscribed
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		numRead := 0
		for env := range f.Chan {
			numRead += 1
			require.Equal(t, contentFilter.ContentTopics[0], env.Message().GetContentTopic())
			t.Log("Received message from the channel")
			wg.Done()
			if numRead == 2 {
				break
			}
		}
	}()

	_, err = node2Relay.PublishToTopic(ctx, tests.CreateWakuMessage(testContentTopic, 0), testTopic)
	require.NoError(t, err)
	_, err = node2Relay.PublishToTopic(ctx, tests.CreateWakuMessage(testContentTopic, 1), testTopic)
	require.NoError(t, err)

	wg.Wait()

	wg.Add(1)
	go func() {
		select {
		case <-f.Chan:
			require.Fail(t, "should not receive another message")
		case <-time.After(1 * time.Second):
			defer wg.Done()
		case <-ctx.Done():
			require.Fail(t, "test exceeded allocated time")
		}
	}()

	// Send to the wrong topic. Should not have anything in the channel
	_, err = node2Relay.PublishToTopic(ctx, tests.CreateWakuMessage("TopicB", 1), testTopic)
	require.NoError(t, err)

	wg.Wait()
}

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Test can't exceed 10 seconds

	testTopic := "/waku/2/go/filter/test"
	testContentTopic := "TopicA"

	filter1, filter2, _ := setupNodes(t, testTopic, ctx)

	contentFilter := &goWakuFilter.ContentFilter{
		Topic:         string(testTopic),
		ContentTopics: []string{testContentTopic},
	}

	_, f, err := filter1.Subscribe(ctx, *contentFilter, goWakuFilter.WithPeer(filter2.h.ID()))
	require.NoError(t, err)

	cancel()
	// Sleep to make sure the filter is subscribed
	time.Sleep(1 * time.Second)

	_, more := <-f.Chan
	if more {
		t.Error("Channel should be closed")
	}
}

// TODO: Add some integration tests with real streams
