package crdt2

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

var visTopicM int
var visTopicN int

func init() {
	flag.IntVar(&visTopicM, "visTopic", 0, "run VisualiseTopic test with specified number of messages")
	flag.IntVar(&visTopicN, "visTopicN", 3, "if running VisualiseTopic test, run with this many nodes")
}

type fixture struct {
	nodes    int
	topics   int
	messages int
}

func Test_RandomMessages(t *testing.T) {
	fixtures := []fixture{
		{5, 1, 100},
		{3, 3, 100},
		{10, 10, 1000},
		{10, 5, 10000},
	}
	if !testing.Short() {
		fixtures = append(fixtures, fixture{30, 1000, 50000}) // should take about 20s locally
	}
	for i, fix := range fixtures {
		t.Run(fmt.Sprintf("%d/%dn/%dt/%dm", i, fix.nodes, fix.topics, fix.messages),
			func(t *testing.T) { randomMsgTest(t, fix.nodes, fix.topics, fix.messages) },
		)
	}
}

func Test_VisualiseTopic(t *testing.T) {
	if visTopicM == 0 {
		return
	}
	net := randomMsgTest(t, visTopicN, 1, visTopicM)
	net.visualiseTopic(os.Stdout, t0)
}

func randomMsgTest(t *testing.T, nodes, topics, messages int) *network {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// to emulate significant concurrent activity we want nodes to be adding
	// events concurrently, but we also want to allow propagation at the same time.
	// So we need to introduce short delays to allow the network
	// to make some propagation progress. Given the random spraying approach
	// injecting a delay at every (nodes*topics)th event should allow most nodes
	// to inject an event to most topics, and then the random length of the delay
	// should allow some amount of propagation to happen before the next burst.
	delayEvery := nodes * topics
	net := NewNetwork(t, ctx, nodes, topics)
	for i := 0; i < messages; i++ {
		topic := fmt.Sprintf("t%d", rand.Intn(topics))
		msg := fmt.Sprintf("gm %d", i)
		net.Publish(rand.Intn(nodes), topic, msg)
		if i%delayEvery == 0 {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		}
	}
	net.AssertEventuallyConsistent(time.Duration(messages*nodes*10) * time.Millisecond)
	return net
}
