package proio // import "github.com/decibelcooper/proio/go-proio"

import (
	"errors"
	"fmt"

	"github.com/decibelcooper/proio/go-proio/proto"
)

// Struct representing an event either created with NewEvent() or retrieved
// with (*Reader) Get() or (*Reader) ScanEvents()
type Event struct {
	header  *proto.EventHeader
	payload []byte

	collCache   map[string]*Collection
	namesCached []string
}

// Returns a new event with minimal initialization
func NewEvent() *Event {
	return &Event{
		header:    &proto.EventHeader{},
		collCache: make(map[string]*Collection),
	}
}

// Get a list of collection names in the event
func (evt *Event) GetNames() []string {
	names := make([]string, 0)

	for _, collHdr := range evt.header.PayloadCollections {
		names = append(names, collHdr.Name)
	}
	for name := range evt.collCache {
		names = append(names, name)
	}

	return names
}

var (
	ErrDupCollection = errors.New("duplicate collection name")
	ErrCollNotFound  = errors.New("collection not found, or zero-length type name")
)

func (evt *Event) SetRunNumber(num uint64) {
	evt.header.RunNumber = num
}

func (evt *Event) SetEventNumber(num uint64) {
	evt.header.EventNumber = num
}

func (evt *Event) GetRunNumber() uint64 {
	return evt.header.RunNumber
}

func (evt *Event) GetEventNumber() uint64 {
	return evt.header.EventNumber
}

func (evt *Event) NewCollection(name, entryType string) (*Collection, error) {
	for key, _ := range evt.collCache {
		if key == name {
			return nil, ErrDupCollection
		}
	}
	for _, collHdr := range evt.header.PayloadCollections {
		if collHdr.Name == name {
			return nil, ErrDupCollection
		}
	}

	coll, err := newCollection(name, entryType, evt.newID())
	if err != nil {
		return nil, err
	}
	evt.collCache[name] = coll
	evt.namesCached = append(evt.namesCached, name)
	return coll, nil
}

// Remove a collection from the event by name
func (evt *Event) Remove(name string) {
	for key := range evt.collCache {
		if key == name {
			delete(evt.collCache, key)
			return
		}
	}
	for _, collHdr := range evt.header.PayloadCollections {
		if collHdr.Name == name {
			evt.getFromPayload(0, name, false)
			return
		}
	}
}

// Gets a collection from the event.  The collection is deserialized upon the
// first time calling this function.  Once deserialized, the collection is
// removed from header.PayloadCollection, and placed back into a queue for
// reserialization.  The event may be safely modified before reserializing.
func (evt *Event) Get(name string) (*Collection, error) {
	if msg := evt.collCache[name]; msg != nil {
		return msg, nil
	}

	return evt.getFromPayload(0, name, true)
}

func (evt *Event) GetEntry(id uint64) Entry {
	collID := uint32(id & 0xffffffff)
	coll, err := evt.getByID(collID)
	if err != nil {
		return nil
	}
	return coll.GetEntry(id)
}

func (evt *Event) String() string {
	output := fmt.Sprintf("Run %v, Event %v\n", evt.GetRunNumber(), evt.GetEventNumber())
	for _, collName := range evt.GetNames() {
		coll, _ := evt.Get(collName)
		output += coll.String()
	}
	return output
}

func (evt *Event) GetHeader() *proto.EventHeader {
	return evt.header
}

func (evt *Event) newID() uint32 {
	evt.header.NUniqueCollIDs++
	return evt.header.NUniqueCollIDs
}

func (evt *Event) getByID(id uint32) (*Collection, error) {
	for _, coll := range evt.collCache {
		if coll.id == id {
			return coll, nil
		}
	}

	return evt.getFromPayload(id, "", true)
}

func (evt *Event) getFromPayload(id uint32, name string, unmarshal bool) (*Collection, error) {
	offset := uint32(0)
	size := uint32(0)
	collName := ""
	collType := ""
	collID := uint32(0)
	collIndex := 0
	var collHdr *proto.EventHeader_CollectionHeader
	for collIndex, collHdr = range evt.header.PayloadCollections {
		if collHdr.Id == id || collHdr.Name == name {
			collName = collHdr.Name
			collType = collHdr.EntryType
			collID = collHdr.Id
			size = collHdr.PayloadSize
			break
		}
		offset += collHdr.PayloadSize
	}
	if collType == "" {
		return nil, ErrCollNotFound
	}

	var coll *Collection
	if unmarshal {
		var err error
		coll, err = newCollection(collName, collType, collID)
		if err != nil {
			return nil, err
		}
		if err = coll.unmarshal(evt.payload[offset : offset+size]); err != nil {
			return nil, err
		}

		evt.collCache[collName] = coll
		evt.namesCached = append(evt.namesCached, collName)
	}

	evt.header.PayloadCollections = append(evt.header.PayloadCollections[:collIndex], evt.header.PayloadCollections[collIndex+1:]...)
	evt.payload = append(evt.payload[:offset], evt.payload[offset+size:]...)

	return coll, nil
}

func (evt *Event) flushCollCache() error {
	for _, name := range evt.namesCached {
		coll := evt.collCache[name]
		if err := evt.collToPayload(coll, name); err != nil {
			return err
		}
		delete(evt.collCache, name)
	}
	evt.namesCached = nil
	return nil
}

func (evt *Event) collToPayload(coll *Collection, name string) error {
	collHdr := &proto.EventHeader_CollectionHeader{
		Name:      name,
		Id:        coll.id,
		EntryType: coll.entryTypeName,
	}

	collBuf, err := coll.marshal()
	if err != nil {
		return err
	}
	collHdr.PayloadSize = uint32(len(collBuf))

	if evt.header == nil {
		evt.header = &proto.EventHeader{}
	}
	evt.header.PayloadCollections = append(evt.header.PayloadCollections, collHdr)
	evt.payload = append(evt.payload, collBuf...)

	return nil
}
