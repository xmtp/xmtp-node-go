package crdt2

import mh "github.com/multiformats/go-multihash"

type Store interface {
	Put(*Event) error
	Get(mh.Multihash) (*Event, error)
	Has(mh.Multihash) (bool, error)
}
