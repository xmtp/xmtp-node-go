package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DispatcherRegistration(t *testing.T) {
	d := newDispatcher()
	sub1 := d.Register(nil, "a", "b", "c")
	sub2 := d.Register(nil, "c", "d")
	sub3 := d.Register(nil, "c", "d", "e", "a")
	requireSubsEqual(t, d.subsByTopic["a"], sub1, sub3)
	requireSubsEqual(t, d.subsByTopic["b"], sub1)
	requireSubsEqual(t, d.subsByTopic["c"], sub1, sub2, sub3)
	requireSubsEqual(t, d.subsByTopic["d"], sub2, sub3)
	requireSubsEqual(t, d.subsByTopic["e"], sub3)

	// unregister unsubscribed topics => noop
	d.Unregister(sub2, "a", "b")
	requireSubsEqual(t, d.subsByTopic["a"], sub1, sub3)
	requireSubsEqual(t, d.subsByTopic["b"], sub1)
	require.Equal(t, 5, len(d.bcsByTopic))
	require.Equal(t, 5, len(d.subsByTopic))
	require.Equal(t, 3, len(d.topicsBySub))

	// unregister topic with 2 subscriptions
	d.Unregister(sub1, "a")
	requireSubsEqual(t, d.subsByTopic["a"], sub3)
	d.Unregister(sub3, "a")
	requireSubsEqual(t, d.subsByTopic["a"])
	require.Equal(t, 4, len(d.bcsByTopic))
	require.Equal(t, 4, len(d.subsByTopic))

	// unregister all subscriptions with leftovers
	d.Unregister(sub2)
	requireSubsEqual(t, d.subsByTopic["c"], sub1, sub3)
	requireSubsEqual(t, d.subsByTopic["d"], sub3)
	require.Equal(t, 4, len(d.bcsByTopic))
	require.Equal(t, 4, len(d.subsByTopic))
	require.Equal(t, 2, len(d.topicsBySub))

	// unregister all subscriptions mixed
	d.Unregister(sub1)
	requireSubsEqual(t, d.subsByTopic["c"], sub3)
	require.Equal(t, 3, len(d.bcsByTopic))
	require.Equal(t, 3, len(d.subsByTopic))
	require.Equal(t, 1, len(d.topicsBySub))

	// unregister last subscription
	d.Unregister(sub3)
	require.Equal(t, 0, len(d.bcsByTopic))
	require.Equal(t, 0, len(d.subsByTopic))
	require.Equal(t, 0, len(d.topicsBySub))
}

func Test_DispatcherUpdates(t *testing.T) {
	d := newDispatcher()
	sub := d.Register(nil, "a", "b", "c")
	require.Equal(t, 3, len(d.subsByTopic))
	require.Equal(t, sub, d.Register(sub, "c", "d", "e"))
	require.Equal(t, 5, len(d.subsByTopic))
	d.Unregister(sub, "a", "b", "e", "x", "y", "z")
	require.Equal(t, 2, len(d.subsByTopic))
	sub2 := d.Register(nil, "d", "e")
	require.Equal(t, sub2, d.Register(sub2, "c", "f"))
	require.Equal(t, 4, len(d.subsByTopic))
	requireSubsEqual(t, d.subsByTopic["c"], sub, sub2)
	requireSubsEqual(t, d.subsByTopic["d"], sub, sub2)
	requireSubsEqual(t, d.subsByTopic["e"], sub2)
	requireSubsEqual(t, d.subsByTopic["f"], sub2)
	d.Unregister(sub2)
	require.Equal(t, 2, len(d.subsByTopic))
	d.Unregister(sub)
	require.Equal(t, 0, len(d.subsByTopic))
}

func Test_DispatcherClose(t *testing.T) {
	d := newDispatcher()
	d.Register(nil, "a", "b", "c")
	d.Register(nil, "c", "d")
	d.Register(nil, "c", "d", "e", "a")
	require.NoError(t, d.Close())
	require.Equal(t, 0, len(d.bcsByTopic))
	require.Equal(t, 0, len(d.subsByTopic))
	require.Equal(t, 0, len(d.topicsBySub))
	require.NoError(t, d.Close())
}

func requireSubsEqual(t *testing.T, subs1 map[chan interface{}]bool, subs2 ...chan interface{}) {
	t.Helper()
	require.Equal(t, len(subs2), len(subs1), "lengths different")
	for _, sub := range subs2 {
		require.True(t, subs1[sub], "sub missing")
	}
}
