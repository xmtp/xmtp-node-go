package crdt2

import (
	"io"

	mh "github.com/multiformats/go-multihash"
)

// Event represents a node in the Merkle-Clock
type Event struct {
	payload []byte
	links   []mh.Multihash // cid's of direct ancestors
	cid     mh.Multihash   // cid is computed by hashing the links and payload together
}

func NewEvent(payload []byte, links []mh.Multihash) (*Event, error) {
	ev := &Event{payload: payload, links: links}
	var err error
	ev.cid, err = mh.SumStream(ev.Reader(), mh.SHA2_256, -1)
	if err != nil {
		return nil, err
	}
	return ev, nil
}

// EventReader helps computing an Event CID efficiently by
// yielding the bytes of the event links first and then the bytes of the payload
// without having to concatenate them all.
// This allows passing the reader to mh.SumStream()
type EventReader struct {
	*Event                     // the event to read from
	unreadLinks []mh.Multihash // copy of event's link slice
	pos         int            // current position from the start of the next link or payload
}

func (ev *Event) Reader() *EventReader {
	return &EventReader{ev, ev.links[:], 0}
}

func (r *EventReader) Read(b []byte) (n int, err error) {
	if len(r.unreadLinks) == 0 && r.pos >= len(r.payload) {
		return 0, io.EOF
	}
	total := 0
	for len(b) > 0 && len(r.unreadLinks) > 0 {
		link := r.unreadLinks[0]
		n := copy(b, link[r.pos:])
		total += n
		b = b[n:]
		r.pos += n
		if r.pos == len(link) {
			r.pos = 0
			r.unreadLinks = r.unreadLinks[1:]
		}
	}
	n = copy(b, r.payload[r.pos:])
	total += n
	r.pos += n
	return total, nil
}
