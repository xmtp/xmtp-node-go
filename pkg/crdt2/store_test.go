package crdt2

import (
	"sync"

	mh "github.com/multiformats/go-multihash"
	messagev1 "github.com/xmtp/proto/v3/go/message_api/v1"
)

type mapStore struct{}

func NewMapStore() *mapStore {
	return &mapStore{}
}

func (s *mapStore) NewTopic(name string) TopicStore {
	return &mapTopicStore{
		heads:  make(map[string]bool),
		events: make(map[string]*Event),
	}
}

// In-memory TopicStore for testing
type mapTopicStore struct {
	sync.Mutex
	heads  map[string]bool   // CIDs of current head events
	events map[string]*Event // maps CIDs to all known Events
}

var _ TopicStore = (*mapTopicStore)(nil)

func (s *mapTopicStore) AddHead(ev *Event) (added bool, err error) {
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

func (s *mapTopicStore) RemoveHead(cid mh.Multihash) (have bool, err error) {
	s.Lock()
	defer s.Unlock()
	key := cid.String()
	if s.events[key] == nil {
		return false, nil
	}
	delete(s.heads, key)
	return true, nil
}

func (s mapTopicStore) NewEvent(env *messagev1.Envelope) (*Event, error) {
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

func (s mapTopicStore) Get(cid mh.Multihash) (*Event, error) {
	return s.events[cid.String()], nil
}

func (s mapTopicStore) allHeads() (cids []mh.Multihash) {
	for key := range s.heads {
		cids = append(cids, s.events[key].cid)
	}
	return cids
}
