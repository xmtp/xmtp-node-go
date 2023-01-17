package crdt2

import (
	"sync"

	mh "github.com/multiformats/go-multihash"
)

type mapStore struct {
	sync.Mutex
	heads  *mh.Set
	events map[string]*Event
}

var _ TopicStore = (*mapStore)(nil)

func NewMapStore() *mapStore {
	return &mapStore{
		heads:  mh.NewSet(),
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
	s.heads.Add(ev.cid)
	return true, nil
}

func (s *mapStore) RemoveHead(cid mh.Multihash) (have bool, err error) {
	s.Lock()
	defer s.Unlock()
	if s.events[cid.String()] == nil {
		return false, nil
	}
	s.heads.Remove(cid)
	return true, nil
}

func (s mapStore) NewEvent(payload []byte) (*Event, error) {
	s.Lock()
	defer s.Unlock()
	ev, err := NewEvent(payload, s.heads.All())
	if err != nil {
		return nil, err
	}
	s.events[ev.cid.String()] = ev
	s.heads = mh.NewSet()
	s.heads.Add(ev.cid)
	return ev, err
}
