package e2e

import (
	"fmt"
	"time"

	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakuprotocol "github.com/status-im/go-waku/waku/v2/protocol"
	wakupb "github.com/status-im/go-waku/waku/v2/protocol/pb"
)

func (e *E2E) testPublishSubscribeQuery() error {
	// Fetch bootstrap node addresses.
	var bootstrapAddrs []string
	if len(e.config.BootstrapAddrs) == 0 {
		var err error
		bootstrapAddrs, err = fetchBootstrapAddrs(e.config.NodesURL, e.config.NetworkEnv)
		if err != nil {
			return err
		}
		if len(bootstrapAddrs) != 3 {
			return fmt.Errorf("expected bootstrap addrs length 3, got: %d", len(bootstrapAddrs))
		}
	} else {
		bootstrapAddrs = e.config.BootstrapAddrs
	}

	// Create a client node for each bootstrap node, and connect to it.
	clients := make([]*wakunode.WakuNode, len(bootstrapAddrs))
	for i, addr := range bootstrapAddrs {
		c, cleanup, err := newNode(
			e.log,
			// Specify libp2p options here to avoid using the waku-default that
			// enables the NAT service, which currently leaks goroutines over
			// time when creating and destroying many in-process.
			// https://github.com/libp2p/go-libp2p/blob/8de2efdb5cfb32daaec7fac71e977761b24be46d/config/config.go#L302
			wakunode.WithLibP2POptions(),
			wakunode.WithoutWakuRelay(),
		)
		if err != nil {
			return err
		}
		defer cleanup()
		err = connectWithAddr(e.ctx, c, addr)
		if err != nil {
			return err
		}
		clients[i] = c
	}
	time.Sleep(500 * time.Millisecond)

	// Subscribe to a topic on each client, connected to each node.
	contentTopic := "test-" + randomStringLower(5)
	envCs := make([]chan *wakuprotocol.Envelope, len(clients))
	for i, c := range clients {
		var err error
		envCs[i], err = subscribeTo(e.ctx, c, []string{contentTopic})
		if err != nil {
			return err
		}
	}
	time.Sleep(500 * time.Millisecond)

	// Send a message to every node.
	msgs := make([]*wakupb.WakuMessage, len(clients))
	for i := range clients {
		msgs[i] = newMessage(contentTopic, int64(i+1), fmt.Sprintf("msg%d", i+1))
	}
	for i, sender := range clients {
		err := publish(e.ctx, sender, msgs[i])
		if err != nil {
			return err
		}
	}

	// Expect them to be relayed to all nodes.
	for _, envC := range envCs {
		err := subscribeExpect(envC, msgs)
		if err != nil {
			return err
		}
	}

	// Expect that they've all been stored on each node.
	for i, c := range clients {
		err := expectQueryMessagesEventually(e.log, c, bootstrapAddrs[i], []string{contentTopic}, msgs)
		if err != nil {
			return err
		}
	}

	return nil
}
