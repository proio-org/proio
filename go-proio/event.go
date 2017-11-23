package proio // import "github.com/decibelcooper/proio/go-proio"

import (
	"errors"
	"fmt"

	"github.com/decibelcooper/proio/go-proio/proto"
	protobuf "github.com/golang/protobuf/proto"
)

// An Event is either created with NewEvent() or retrieved with (*Reader) Get()
// or (*Reader) ScanEvents().
type Event struct {
	proto *proto.EventProto
}

// NewEvent returns a new event with minimal initialization.
func NewEvent() *Event {
	return &Event{
		proto: &proto.EventProto{
			collections: make(map[uint32]*proto.CollectionProto),
		},
	}
}

//func (evt *Event) NewCollection(name, entryType string)
//
//func (evt *Event) CollIDs(sorted bool) []uint64 {
//	var ids []uint64
//	for id, _ := range coll.entryCache {
//		ids = append(ids, uint64(id)<<32+uint64(coll.id))
//	}
//	for id, _ := range coll.proto.Entries {
//		ids = append(ids, uint64(id)<<32+uint64(coll.id))
//	}
//	if sorted {
//		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
//	}
//	return ids
//}
