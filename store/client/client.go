package client

import (
	"context"
	"encoding/hex"
	"math"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-msgio/protoio"
	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/xmtp/xmtp-node-go/logging"
	"github.com/xmtp/xmtp-node-go/metrics"
	"go.uber.org/zap"
)

type Client struct {
	log  *zap.Logger
	host host.Host
	peer *peer.ID
}

func New(opts ...ClientOption) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	// Required logger option.
	if c.log == nil {
		return nil, errors.New("missing logger option")
	}
	c.log = c.log.Named("client")

	// Required host option.
	if c.host == nil {
		return nil, errors.New("missing host option")
	}
	c.log = c.log.With(zap.String("host", c.host.ID().Pretty()))

	// Required peer option.
	if c.peer == nil {
		return nil, errors.New("missing peer option")
	}
	c.log = c.log.With(zap.String("peer", c.peer.Pretty()))

	return c, nil
}

// Query executes query requests against a peer, calling the given pageFn with
// every page response, traversing every page until the end or until pageFn
// returns false.
func (c *Client) Query(ctx context.Context, query *pb.HistoryQuery, pageFn func(res *pb.HistoryResponse) (int, bool)) (int, error) {
	var msgCount int
	var msgCountLock sync.RWMutex
	var res *pb.HistoryResponse
	for {
		if res != nil {
			if res.PagingInfo == nil || res.PagingInfo.Cursor == nil || len(res.Messages) == 0 {
				break
			}
			query.PagingInfo = res.PagingInfo
		}
		var err error
		res, err = c.queryFrom(ctx, query, protocol.GenerateRequestId())
		if err != nil {
			return 0, err
		}
		count, ok := pageFn(res)
		if !ok {
			break
		}
		msgCountLock.Lock()
		msgCount += count
		msgCountLock.Unlock()
	}
	return msgCount, nil
}

func (c *Client) queryFrom(ctx context.Context, q *pb.HistoryQuery, requestId []byte) (*pb.HistoryResponse, error) {
	peer := *c.peer
	c.log.Info("querying from peer", logging.HostID("peer", peer))

	// We connect first so dns4 addresses are resolved (NewStream does not do it)
	err := c.host.Connect(ctx, c.host.Peerstore().PeerInfo(peer))
	if err != nil {
		return nil, err
	}

	stream, err := c.host.NewStream(ctx, peer, store.StoreID_v20beta4)
	if err != nil {
		c.log.Error("connecting to peer", zap.Error(err))
		return nil, err
	}
	defer stream.Close()
	defer func() {
		_ = stream.Reset()
	}()

	historyRequest := &pb.HistoryRPC{Query: q, RequestId: hex.EncodeToString(requestId)}
	writer := protoio.NewDelimitedWriter(stream)
	err = writer.WriteMsg(historyRequest)
	if err != nil {
		c.log.Error("writing request", zap.Error(err))
		return nil, err
	}

	historyResponseRPC := &pb.HistoryRPC{}
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)
	err = reader.ReadMsg(historyResponseRPC)
	if err != nil {
		c.log.Error("reading response", zap.Error(err))
		metrics.RecordStoreError(ctx, "decodeRPCFailure")
		return nil, err
	}

	metrics.RecordMessage(ctx, "retrieved", 1)
	return historyResponseRPC.Response, nil
}
