package crdt

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	proto "github.com/xmtp/proto/v3/go/message_api/v1"
)

const (
	envelopesKeyNamespace = "envelopes"
)

func buildMessageQueryPrefix(topic string) string {
	return strings.Join([]string{
		envelopesKeyNamespace,
		encodeTopicForStoreKey(topic),
	}, "/") + "/"
}

func buildMessageStoreKey(env *proto.Envelope) (datastore.Key, error) {
	cID, err := newCID(env.Message)
	if err != nil {
		return datastore.Key{}, errors.Wrap(err, "creating cid")
	}

	key := datastore.NewKey(strings.Join([]string{
		envelopesKeyNamespace,
		encodeTopicForStoreKey(env.ContentTopic),
		fmt.Sprintf("%020d", env.TimestampNs),
		cID.String(),
	}, "/"))
	return key, nil
}

func encodeTopicForStoreKey(topic string) string {
	return base64.StdEncoding.EncodeToString([]byte(topic))
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
