package e2e

import (
	"errors"
	"fmt"
	"time"

	wakunode "github.com/status-im/go-waku/waku/v2/node"
	wakuprotocol "github.com/status-im/go-waku/waku/v2/protocol"
	wakupb "github.com/status-im/go-waku/waku/v2/protocol/pb"
	"go.uber.org/zap"
)

type wakuClient struct {
	addr string
	node *wakunode.WakuNode
}

func (s *Suite) testWakuPublishSubscribeQuery(log *zap.Logger) error {
	// Fetch bootstrap node addresses.
	var bootstrapAddrs []string
	if len(s.config.BootstrapAddrs) == 0 {
		var err error
		bootstrapAddrs, err = fetchBootstrapAddrs(s.config.NodesURL, s.config.NetworkEnv)
		if err != nil {
			return err
		}
		if len(bootstrapAddrs) != 3 {
			return fmt.Errorf("expected bootstrap addrs length 3, got: %d", len(bootstrapAddrs))
		}
	} else {
		bootstrapAddrs = s.config.BootstrapAddrs
	}

	// Create a client node for each bootstrap node, and connect to it.
	clients := make([]*wakuClient, 0, len(bootstrapAddrs))
	for _, addr := range bootstrapAddrs {
		c, cleanup, err := newNode(
			log,
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
		err = wakuConnectWithAddr(s.ctx, c, addr)
		if err != nil {
			log.Info("ignoring error while connecting to bootstrap node", zap.Error(err))
			continue
		}
		clients = append(clients, &wakuClient{
			addr: addr,
			node: c,
		})
	}
	if len(clients) == 0 {
		return errors.New("no bootstrap nodes available")
	}
	time.Sleep(500 * time.Millisecond)

	// Subscribe to a topic on each client, connected to each node.
	contentTopic := "test-" + s.randomStringLower(12)
	envCs := make([]chan *wakuprotocol.Envelope, len(clients))
	for i, c := range clients {
		var err error
		envCs[i], err = wakuSubscribeTo(s.ctx, c.node, []string{contentTopic})
		if err != nil {
			return err
		}
	}
	time.Sleep(500 * time.Millisecond)

	// Send a message to every node.
	msgs := make([]*wakupb.WakuMessage, len(clients))
	for i := range clients {
		msgs[i] = newWakuMessage(contentTopic, int64(i+1), fmt.Sprintf("msg%d", i+1))
	}
	for i, sender := range clients {
		err := wakuPublish(s.ctx, sender.node, msgs[i])
		if err != nil {
			return err
		}
	}

	// Expect them to be relayed to all nodes.
	for _, envC := range envCs {
		err := wakuSubscribeExpect(envC, msgs)
		if err != nil {
			return err
		}
	}

	// Expect that they've all been stored on each node.
	for i, c := range clients {
		err := wakuExpectQueryMessagesEventually(log, c.node, bootstrapAddrs[i], []string{contentTopic}, msgs)
		if err != nil {
			return err
		}
	}

	return nil
}
