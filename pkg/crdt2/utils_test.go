package crdt2

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
	test "github.com/xmtp/xmtp-node-go/pkg/testing"
	"go.uber.org/zap"
)

// network is an in-memory simulation of a network of a given number of Nodes.
// All nodes host all of the given number of topics named t0, t1, etc.
// network also captures events that were published to it for final analysis of the test results.
type network struct {
	t      *testing.T
	nodes  []*Node  // all the nodes in the network
	events []*Event // captures events published to the network
}

const t0 = "t0" // first topic
// const t1 = "t1" // second topic
// const t2 = "t2" // third topic
// ...

// Creates a network with given number of nodes and given number of topics on all nodes
func NewNetwork(t *testing.T, ctx context.Context, nodes, topics int) *network {
	log := test.NewLog(t)
	bc := NewChanBroadcaster(log)
	sync := NewRandomSyncer()
	var list []*Node
	for i := 0; i < nodes; i++ {
		name := fmt.Sprintf("n%d", i)
		n := NewNode(ctx,
			log.Named(name),
			NewMapStore(),
			sync,
			bc)
		bc.AddNode(n)
		sync.AddNode(n)
		for j := 0; j < topics; j++ {
			topic := fmt.Sprintf("t%d", j)
			n.NewTopic(topic)
			log.Debug("creating", zap.String("node", name), zap.String("topic", topic))
		}
		require.Len(t, n.Topics, topics)
		list = append(list, n)
	}
	require.Len(t, list, nodes)
	require.Len(t, bc.subscribers, nodes)
	require.Len(t, sync.nodes, nodes)
	return &network{t: t, nodes: list}
}

// Publishes msg into a topic from given node
func (net *network) Publish(node int, topic, msg string) {
	net.t.Helper()
	n := net.nodes[node]
	ev, err := n.Publish(n.ctx, &messagev1.Envelope{TimestampNs: uint64(len(net.events) + 1), ContentTopic: topic, Message: []byte(msg)})
	assert.NoError(net.t, err)
	net.events = append(net.events, ev)
}

// Suspends topic broadcast delivery to the given node while fn runs
func (net *network) WithSuspendedTopic(node int, topic string, fn func(*Node)) {
	n := net.nodes[node]
	bc := n.NodeBroadcaster.(*ChanBroadcaster)
	bc.RemoveNode(n)
	defer bc.AddNode(n)
	fn(n)
}

// Wait for all the network nodes to converge on the captured set of events.
func (net *network) AssertEventuallyConsistent(timeout time.Duration, ignore ...int) {
	net.t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timer.C:
			missing := net.checkEvents(ignore)
			if len(missing) > 0 {
				net.t.Errorf("Missing events: %v", missing)
			}
			return
		case <-ticker.C:
			if len(net.checkEvents(ignore)) == 0 {
				return
			}
		}
	}
}

// Check that all nodes except the ignored ones have all events.
// Returns map of nodes that have missing events,
// the key is the node number
// the value is a string listing present events by number and _ for missing events.
func (net *network) checkEvents(ignore []int) (missing map[int]string) {
	missing = make(map[int]string)
	for j, n := range net.nodes {
		if ignored(j, ignore) {
			continue
		}
		result := ""
		pass := true
		for i, ev := range net.events {
			ev2, err := n.Get(ev.ContentTopic, ev.cid)
			if err != nil || ev2 == nil {
				result = result + "_"
				pass = false
			} else {
				result = result + strconv.FormatInt(int64(i), 36)
			}
		}
		if !pass {
			missing[j] = result
		}
	}
	return missing
}

// emit a graphvis depiction of the topic contents
// showing the individual events and their links
func (net *network) visualiseTopic(w io.Writer, topic string) {
	fmt.Fprintf(w, "strict digraph %s {\n", topic)
	for i := len(net.events) - 1; i >= 0; i-- {
		ev := net.events[i]
		if ev.ContentTopic != topic {
			continue
		}
		fmt.Fprintf(w, "\t\"%s\" [label=\"%d: \\N\"]\n", shortenedCid(ev.cid), i)
		fmt.Fprintf(w, "\t\"%s\" -> { ", shortenedCid(ev.cid))
		for _, l := range ev.links {
			fmt.Fprintf(w, "\"%s\" ", shortenedCid(l))
		}
		fmt.Fprintf(w, "}\n")
	}
	fmt.Fprintf(w, "}\n")
}

func ignored(i int, ignore []int) bool {
	for _, j := range ignore {
		if i == j {
			return true
		}
	}
	return false
}
