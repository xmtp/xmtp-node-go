package filter

import (
	"testing"

	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/stretchr/testify/require"
)

func TestFilterMap(t *testing.T) {
	fmap := NewFilterMap()

	filter := Filter{
		PeerID:         "id",
		Topic:          "test",
		ContentFilters: []string{"test"},
		Chan:           make(chan *protocol.Envelope),
	}

	fmap.Set("test", filter)
	res := <-fmap.Items()
	require.Equal(t, "test", res.Key)

	item, ok := fmap.Get("test")
	require.True(t, ok)
	require.Equal(t, "test", item.Topic)

	fmap.Delete("test")

	_, ok = fmap.Get("test")
	require.False(t, ok)
}
