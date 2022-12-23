package crdt

import (
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	proto "github.com/xmtp/proto/v3/go/message_api/v1"
)

const (
	envelopesKeyNamespace = "envelopes"
)

func buildMessageQuery(req *proto.QueryRequest) query.Query {
	var topic string
	if len(req.ContentTopics) > 0 {
		topic = req.ContentTopics[0] // TODO
	}
	q := query.Query{Prefix: buildMessageQueryPrefix(topic, 0)}
	// TODO: sorting, start/end time filtering
	if req.PagingInfo != nil {
		if req.PagingInfo.Direction == proto.SortDirection_SORT_DIRECTION_DESCENDING {
			q.Orders = []query.Order{query.OrderByKeyDescending{}}
		}
		if limit := req.PagingInfo.Limit; limit > 0 {
			q.Limit = int(limit)
		}
	}
	if req.StartTimeNs > 0 {
		q.Filters = append(q.Filters, query.FilterKeyCompare{
			Op:  query.GreaterThanOrEqual,
			Key: buildMessageQueryPrefix(topic, req.StartTimeNs),
		})
	}
	return q
}

func buildMessageQueryPrefix(topic string, timestampNs uint64) string {
	segments := []string{
		envelopesKeyNamespace,
		topic,
	}
	if timestampNs > 0 {
		segments = append(segments, fmt.Sprintf("%020d", timestampNs))
	}
	segments = append(segments, "") // to get a slash at the end
	return strings.Join(segments, "/")
}

func buildMessageStoreKey(env *proto.Envelope) (datastore.Key, error) {
	cID, err := newCID(env.Message)
	if err != nil {
		return datastore.Key{}, errors.Wrap(err, "creating cid")
	}

	key := datastore.NewKey(
		buildMessageQueryPrefix(env.ContentTopic, env.TimestampNs) + cID.String())
	return key, nil
}

func newCID(val []byte) (cid.Cid, error) {
	pref := cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.Raw),
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}
	cID, err := pref.Sum(val)
	if err != nil {
		return cid.Cid{}, err
	}
	return cID, nil
}
