package client

import (
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/zap"
)

type ClientOption func(c *Client)

func WithLog(log *zap.Logger) ClientOption {
	return func(c *Client) {
		c.log = log
	}
}

func WithHost(host host.Host) ClientOption {
	return func(c *Client) {
		c.host = host
	}
}

func WithPeer(id peer.ID) ClientOption {
	return func(c *Client) {
		c.peer = &id
	}
}
