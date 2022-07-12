package store

import (
	"context"
	"encoding/hex"
	"math"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-msgio/protoio"
	"github.com/pkg/errors"
	"github.com/status-im/go-waku/waku/v2/protocol"
	"github.com/status-im/go-waku/waku/v2/protocol/pb"
	"github.com/status-im/go-waku/waku/v2/protocol/store"
	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"go.uber.org/zap"
)

type Client struct {
	log  *zap.Logger
	host host.Host
	peer *peer.ID
}

var (
	ErrMissingLogOption  = errors.New("missing log option")
	ErrMissingHostOption = errors.New("missing host option")
	ErrMissingPeerOption = errors.New("missing peer option")
)

type ClientOption func(c *Client)

func WithClientLog(log *zap.Logger) ClientOption {
	return func(c *Client) {
		c.log = log
	}
}

func WithClientHost(host host.Host) ClientOption {
	return func(c *Client) {
		c.host = host
	}
}

func WithClientPeer(id peer.ID) ClientOption {
	return func(c *Client) {
		c.peer = &id
	}
}

func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	// Required logger option.
	if c.log == nil {
		return nil, ErrMissingLogOption
	}
	c.log = c.log.Named("client")

	// Required host option.
	if c.host == nil {
		return nil, ErrMissingHostOption
	}
	c.log = c.log.With(zap.String("host", c.host.ID().Pretty()))

	// Required peer option.
	if c.peer == nil {
		return nil, ErrMissingPeerOption
	}
	c.log = c.log.With(zap.String("peer", c.peer.Pretty()))

	return c, nil
}

// Query executes query requests against a peer, calling the given pageFn with
// every page response, traversing every page until the end or until pageFn
// returns false.
func (c *Client) Query(ctx context.Context, query *pb.HistoryQuery, pageFn func(res *pb.HistoryResponse) (int, bool)) (int, error) {
	var msgCount int
	var res *pb.HistoryResponse
	for {
		if res != nil {
			if isLastPage(res) {
				break
			}
			query.PagingInfo = res.PagingInfo
		}

		var err error
		res, err = c.queryPage(ctx, query)
		if err != nil {
			return 0, err
		}

		count, ok := pageFn(res)
		if !ok {
			break
		}

		msgCount += count
	}
	return msgCount, nil
}

func (c *Client) queryPage(ctx context.Context, query *pb.HistoryQuery) (*pb.HistoryResponse, error) {
	stream, err := c.host.NewStream(ctx, *c.peer, store.StoreID_v20beta4)
	if err != nil {
		return nil, errors.Wrap(err, "opening query stream")
	}
	defer func() {
		_ = stream.Reset()
		stream.Close()
	}()

	writer := protoio.NewDelimitedWriter(stream)
	err = writer.WriteMsg(&pb.HistoryRPC{
		Query:     query,
		RequestId: hex.EncodeToString(protocol.GenerateRequestId()),
	})
	if err != nil {
		return nil, errors.Wrap(err, "writing query request")
	}

	res := &pb.HistoryRPC{}
	reader := protoio.NewDelimitedReader(stream, math.MaxInt32)
	err = reader.ReadMsg(res)
	if err != nil {
		metrics.RecordStoreError(ctx, "decodeRPCFailure")
		return nil, errors.Wrap(err, "reading query response")
	}

	metrics.RecordMessage(ctx, "retrieved", 1)
	return res.Response, nil
}

func isLastPage(res *pb.HistoryResponse) bool {
	return res.PagingInfo == nil || res.PagingInfo.Cursor == nil || len(res.Messages) == 0
}
