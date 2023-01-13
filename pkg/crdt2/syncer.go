package crdt2

import mh "github.com/multiformats/go-multihash"

type Syncer interface {
	Fetch([]mh.Multihash) ([]*Event, error)
	// FetchAll() []Event // used to seed new nodes
}
