package crdt2

import (
	"crypto/rand"
	"testing"

	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
)

func Test_NilEvent(t *testing.T) {
	ev, err := NewEvent(nil, nil)
	assert.NoError(t, err)
	emptyHash, err := mh.Sum(nil, mh.SHA2_256, -1)
	assert.NoError(t, err)
	assert.Equal(t, ev.cid, emptyHash)
}

func Test_NoLinksEvent(t *testing.T) {
	payload := make([]byte, 1000)
	_, err := rand.Reader.Read(payload)
	assert.NoError(t, err)
	ev, err := NewEvent(payload, nil)
	assert.NoError(t, err)
	hash, err := mh.Sum(payload, mh.SHA2_256, -1)
	assert.NoError(t, err)
	assert.Equal(t, ev.cid, hash)
}

func Test_Event(t *testing.T) {
	payload := make([]byte, 1000)
	_, err := rand.Reader.Read(payload)
	assert.NoError(t, err)
	links := makeLinks(t, "one", "two", "three")
	ev, err := NewEvent(payload, links)
	assert.NoError(t, err)
	hash, err := mh.Sum(join(links, payload), mh.SHA2_256, -1)
	assert.NoError(t, err)
	assert.Equal(t, ev.cid, hash)
}

func makeLinks(t *testing.T, payloads ...string) (links []mh.Multihash) {
	for _, p := range payloads {
		ev, err := NewEvent([]byte(p), nil)
		assert.NoError(t, err)
		links = append(links, ev.cid)
	}
	return links
}

func join(hashes []mh.Multihash, payload []byte) (total []byte) {
	for _, h := range hashes {
		total = append(total, h...)
	}
	return append(total, payload...)
}
