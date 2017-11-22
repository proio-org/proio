package proio // import "github.com/decibelcooper/proio/go-proio"

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/decibelcooper/proio/go-proio/proto"
	protobuf "github.com/golang/protobuf/proto"
)

// A Collection manages serialization and deserialization of entry messages.
// Additionally, a Collection manages reference ids for entries contained
// within it.  Reference ids are persistent through serialization and can be
// used within entry messages to refer to other entries in any collection in
// the event by calling GetEntry on either the collection or the event.
// Collections should be created by calling NewCollection on an event.
type Collection struct {
	Name string

	id            uint32
	proto         *proto.CollectionProto
	entryTypeName string
	entryType     reflect.Type
	entryCache    map[uint32]protobuf.Message
}

var (
	ErrUnknownType  = errors.New("unknown entry type")
	ErrTypeMismatch = errors.New("entry type does not match collection")
	ErrIDMismatch   = errors.New("entry ID does not match collection")
)

// GetType gets a string representing the entry type
func GetType(entry protobuf.Message) string {
	return protobuf.MessageName(entry)
}

// EntryIDs returns a slice of IDs for entries contained in the collection.  The
// boolean argument determines whether or not EntryIDs takes the time to sort
// the slice by ID number before returning.
func (coll *Collection) EntryIDs(sorted bool) []uint64 {
	var ids []uint64
	for id, _ := range coll.entryCache {
		ids = append(ids, uint64(id)<<32+uint64(coll.id))
	}
	for id, _ := range coll.proto.Entries {
		ids = append(ids, uint64(id)<<32+uint64(coll.id))
	}
	if sorted {
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	}
	return ids
}

// GetEntry returns the entry requested by its reference id.  If the id does
// not belong to the collection, nil is returned.
func (coll *Collection) GetEntry(id uint64) protobuf.Message {
	if uint32(id&0xffffffff) != coll.id {
		return nil
	}

	entryID := uint32(id >> 32)
	if entry, ok := coll.entryCache[entryID]; ok {
		return entry
	}

	if entryBytes, ok := coll.proto.Entries[entryID]; ok {
		entry := reflect.New(coll.entryType).Interface().(protobuf.Message)
		selfSerializingEntry, ok := entry.(selfSerializingEntry)
		if ok {
			selfSerializingEntry.Unmarshal(entryBytes)
		} else {
			protobuf.Unmarshal(entryBytes, entry)
		}

		delete(coll.proto.Entries, entryID)
		coll.entryCache[entryID] = entry

		return entry
	}

	return nil
}

// AddEntry adds the given entry to the collection and returns a reference id
// for the entry and an error value.  If the entry type does not match the
// collection type, ErrTypeMismatch is returned (otherwise nil).
func (coll *Collection) AddEntry(entry protobuf.Message) (uint64, error) {
	if reflect.TypeOf(entry).Elem() != coll.entryType {
		return 0, ErrTypeMismatch
	}

	entryID := coll.newID()
	coll.entryCache[entryID] = entry
	return (uint64(entryID)<<32 + uint64(coll.id)), nil
}

// AddEntries performs the same task as AddEntry, except that it accepts an
// arbitrary number of entries and returns a slice of corresponding ids instead
// of just one.
func (coll *Collection) AddEntries(entries ...protobuf.Message) ([]uint64, error) {
	var entryIDs []uint64
	for _, entry := range entries {
		id, err := coll.AddEntry(entry)
		if err != nil {
			return nil, err
		}
		entryIDs = append(entryIDs, id)
	}
	return entryIDs, nil
}

// RemoveEntry removes an entry from the collection by its reference id.
func (coll *Collection) RemoveEntry(id uint64) error {
	if uint32(id&0xffffffff) != coll.id {
		return ErrIDMismatch
	}

	entryID := uint32(id >> 32)
	delete(coll.entryCache, entryID)
	delete(coll.proto.Entries, entryID)
	return nil
}

// AuxData returns a mutable map of auxiliary data that is serialized with the
// collection.  See also the AuxData receiver for Events.
func (coll *Collection) AuxData() map[string][]byte {
	if coll.proto.AuxData == nil {
		coll.proto.AuxData = make(map[string][]byte)
	}
	return coll.proto.AuxData
}

func (coll *Collection) String() string {
	output := fmt.Sprintf("Collection: %s, Type: %s, ID: %v\n", coll.Name, coll.entryTypeName, coll.id)
	for _, entryID := range coll.EntryIDs(true) {
		output += fmt.Sprintf("ID:%v %s\n", entryID, coll.GetEntry(entryID))
	}
	return output
}

type selfSerializingEntry interface {
	protobuf.Message

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func newCollection(name, entryType string, id uint32) (*Collection, error) {
	typeName := entryType
	ptrType := protobuf.MessageType(typeName)
	if ptrType == nil {
		fmt.Println(name, typeName, id)
		return nil, ErrUnknownType
	}

	return &Collection{
		Name:          name,
		id:            id,
		proto:         &proto.CollectionProto{Entries: make(map[uint32][]byte)},
		entryTypeName: entryType,
		entryType:     ptrType.Elem(),
		entryCache:    make(map[uint32]protobuf.Message),
	}, nil
}

func (coll *Collection) NEntries() int {
	return len(coll.entryCache) + len(coll.proto.Entries)
}

func (coll *Collection) unmarshal(bytes []byte) error {
	return coll.proto.Unmarshal(bytes)
}

func (coll *Collection) marshal() ([]byte, error) {
	for id, entry := range coll.entryCache {
		var err error

		selfSerializingEntry, ok := entry.(selfSerializingEntry)
		if ok {
			coll.proto.Entries[id], err = selfSerializingEntry.Marshal()
		} else {
			coll.proto.Entries[id], err = protobuf.Marshal(entry)
		}

		if err != nil {
			return nil, err
		}
	}
	coll.entryCache = make(map[uint32]protobuf.Message)

	return coll.proto.Marshal()
}

func (coll *Collection) newID() uint32 {
	coll.proto.NUniqueEntryIDs++
	return coll.proto.NUniqueEntryIDs
}
