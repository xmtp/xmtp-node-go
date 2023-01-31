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
// network also captures events that were published to it for final analysis of the test results.
type network struct {
	t      *testing.T
	ctx    context.Context
	Cancel context.CancelFunc
	log    *zap.Logger
	bc     *chanBroadcaster
	sync   *randomSyncer
	count  int
	nodes  map[int]*Node // all the nodes in the network
	events []*Event      // captures events published to the network
}

const t0 = "t0" // first topic
// const t1 = "t1" // second topic
// const t2 = "t2" // third topic
// ...

// Creates a network with given number of nodes
func newNetwork(t *testing.T, nodes, topics int) *network {
	ctx, cancel := context.WithCancel(context.Background())
	log := test.NewLog(t)
	net := &network{
		t:      t,
		ctx:    ctx,
		Cancel: cancel,
		log:    log,
		bc:     newChanBroadcaster(log),
		sync:   newRandomSyncer(),
		nodes:  make(map[int]*Node),
	}
	for i := 0; i < nodes; i++ {
		net.AddNode(newMapStore())
	}
	require.Len(t, net.nodes, nodes)
	require.Len(t, net.bc.subscribers, nodes)
	require.Len(t, net.sync.nodes, nodes)
	return net
}

func (net *network) AddNode(store NodeStore) *Node {
	name := fmt.Sprintf("n%d", net.count)
	n, err := NewNode(net.ctx,
		net.log.Named(name),
		store,
		net.sync,
		net.bc)
	assert.NoError(net.t, err)
	net.bc.AddNode(n)
	net.sync.AddNode(n)
	net.nodes[net.count] = n
	net.count += 1
	return n
}

func (net *network) RemoveNode(n int) *Node {
	node := net.nodes[n]
	if node == nil {
		return nil
	}
	delete(net.nodes, n)
	node.Cancel()
	return node
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
	bc := n.NodeBroadcaster.(*chanBroadcaster)
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
		count, err := n.Count()
		assert.NoError(net.t, err)
		if count == len(net.events) {
			continue // shortcut
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
		assert.False(net.t, pass)
		missing[j] = result
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
