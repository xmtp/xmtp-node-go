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
	assertSubsEqual(t, d.subsByTopic["a"], sub1, sub3)
	assertSubsEqual(t, d.subsByTopic["b"], sub1)
	assertSubsEqual(t, d.subsByTopic["c"], sub1, sub2, sub3)
	assertSubsEqual(t, d.subsByTopic["d"], sub2, sub3)
	assertSubsEqual(t, d.subsByTopic["e"], sub3)

	// unregister unsubscribed topics => noop
	d.Unregister(sub2, "a", "b")
	assertSubsEqual(t, d.subsByTopic["a"], sub1, sub3)
	assertSubsEqual(t, d.subsByTopic["b"], sub1)
	require.Equal(t, 5, len(d.bcsByTopic))
	require.Equal(t, 5, len(d.subsByTopic))
	require.Equal(t, 3, len(d.topicsBySub))

	// unregister topic with 2 subscriptions
	d.Unregister(sub1, "a")
	assertSubsEqual(t, d.subsByTopic["a"], sub3)
	d.Unregister(sub3, "a")
	assertSubsEqual(t, d.subsByTopic["a"])
	require.Equal(t, 4, len(d.bcsByTopic))
	require.Equal(t, 4, len(d.subsByTopic))

	// unregister all subscriptions with leftovers
	d.Unregister(sub2)
	assertSubsEqual(t, d.subsByTopic["c"], sub1, sub3)
	assertSubsEqual(t, d.subsByTopic["d"], sub3)
	require.Equal(t, 4, len(d.bcsByTopic))
	require.Equal(t, 4, len(d.subsByTopic))
	require.Equal(t, 2, len(d.topicsBySub))

	// unregister all subscriptions mixed
	d.Unregister(sub1)
	assertSubsEqual(t, d.subsByTopic["c"], sub3)
	require.Equal(t, 3, len(d.bcsByTopic))
	require.Equal(t, 3, len(d.subsByTopic))
	require.Equal(t, 1, len(d.topicsBySub))

	// unregister last subscription
	d.Unregister(sub3)
	require.Equal(t, 0, len(d.bcsByTopic))
	require.Equal(t, 0, len(d.subsByTopic))
	require.Equal(t, 0, len(d.topicsBySub))
}

func assertSubsEqual(t *testing.T, subs1 map[chan interface{}]bool, subs2 ...chan interface{}) {
	t.Helper()
	require.Equal(t, len(subs2), len(subs1), "lengths different")
	for _, sub := range subs2 {
		require.True(t, subs1[sub], "sub missing")
	}
}
