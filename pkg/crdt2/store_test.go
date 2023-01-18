package crdt2

import (
	"sync"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

// In-memory TopicStore for testing
type mapStore struct {
	sync.Mutex
	heads  map[string]bool   // CIDs of current head events
	events map[string]*Event // maps CIDs to all known Events
}

var _ TopicStore = (*mapStore)(nil)

func NewMapStore() *mapStore {
	return &mapStore{
		heads:  make(map[string]bool),
		events: make(map[string]*Event),
	}
}

func (s *mapStore) AddHead(ev *Event) (added bool, err error) {
	s.Lock()
	defer s.Unlock()
	key := ev.cid.String()
	if s.events[key] != nil {
		return false, nil
	}
	s.events[key] = ev
	s.heads[key] = true
	return true, nil
}

func (s *mapStore) RemoveHead(cid mh.Multihash) (have bool, err error) {
	s.Lock()
	defer s.Unlock()
	key := cid.String()
	if s.events[key] == nil {
		return false, nil
	}
	delete(s.heads, key)
	return true, nil
}

func (s mapStore) NewEvent(env *messagev1.Envelope) (*Event, error) {
	s.Lock()
	defer s.Unlock()
	ev, err := NewEvent(env, s.allHeads())
	if err != nil {
		return nil, err
	}
	key := ev.cid.String()
	s.events[key] = ev
	s.heads = map[string]bool{key: true}
	return ev, err
}

func (s mapStore) allHeads() (cids []mh.Multihash) {
	for key := range s.heads {
		cids = append(cids, s.events[key].cid)
	}
	return cids
}
