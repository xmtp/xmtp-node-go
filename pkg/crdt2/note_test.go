package crdt2

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test_RandomMessages(t *testing.T) {
	for i, fix := range []struct {
		nodes    int
		topics   int
		messages int
	}{
		{5, 1, 100},
		{3, 3, 100},
		{10, 10, 1000},
		{10, 5, 10000},
	} {
		t.Run(fmt.Sprintf("%d/%dn/%dt/%dm", i, fix.nodes, fix.topics, fix.messages),
			func(t *testing.T) { randomMsgTest(t, fix.nodes, fix.topics, fix.messages) },
		)
	}
}

func randomMsgTest(t *testing.T, nodes, topics, messages int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	net := NewNetwork(t, ctx, nodes, topics)
	for i := 0; i < messages; i++ {
		topic := fmt.Sprintf("t%d", rand.Intn(topics))
		msg := fmt.Sprintf("gm %d", i)
		net.Publish(rand.Intn(nodes), topic, msg)
		if i%5 == 0 {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		}
	}
	net.AssertEventuallyConsistent(time.Duration(messages*nodes*10) * time.Millisecond)
}
