package proio // import "github.com/decibelcooper/proio/go-proio"

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/decibelcooper/proio/go-proio/proto"
	protobuf "github.com/golang/protobuf/proto"
)

type Entry interface {
	protobuf.Message

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type idSlice []uint64

type Collection struct {
	Name string

	id            uint32
	proto         *proto.CollectionProto
	entryTypeName string
	entryType     reflect.Type
	entryCache    map[uint32]Entry
}

var (
	ErrUnknownType  = errors.New("unknown entry type")
	ErrTypeMismatch = errors.New("entry type does not match collection")
	ErrIDMismatch   = errors.New("entry ID does not match collection")
)

// Get a string representing the entry type
func GetType(entry Entry) string {
	return strings.TrimPrefix(protobuf.MessageName(entry), "proio.model.")
}

func newCollection(name, entryType string, id uint32) (*Collection, error) {
	typeName := "proio.model." + entryType
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
		entryCache:    make(map[uint32]Entry),
	}, nil
}

func (coll *Collection) NEntries() int {
	return len(coll.entryCache) + len(coll.proto.Entries)
}

func (coll *Collection) EntryIDs(sorted bool) []uint64 {
	var ids idSlice
	for id, _ := range coll.entryCache {
		ids = append(ids, uint64(id)<<32+uint64(coll.id))
	}
	for id, _ := range coll.proto.Entries {
		ids = append(ids, uint64(id)<<32+uint64(coll.id))
	}
	if sorted {
		sort.Sort(ids)
	}
	return ids
}

func (coll *Collection) GetEntry(id uint64) Entry {
	if uint32(id&0xffffffff) != coll.id {
		return nil
	}

	entryID := uint32(id >> 32)
	if entry, ok := coll.entryCache[entryID]; ok {
		return entry
	}

	if entryBytes, ok := coll.proto.Entries[entryID]; ok {
		entry := reflect.New(coll.entryType).Interface().(Entry)
		entry.Unmarshal(entryBytes)

		delete(coll.proto.Entries, entryID)
		coll.entryCache[entryID] = entry

		return entry
	}

	return nil
}

func (coll *Collection) AddEntry(entry Entry) (uint64, error) {
	if reflect.TypeOf(entry).Elem() != coll.entryType {
		return 0, ErrTypeMismatch
	}

	entryID := coll.newID()
	coll.entryCache[entryID] = entry
	return (uint64(entryID)<<32 + uint64(coll.id)), nil
}

func (coll *Collection) AddEntries(entries ...Entry) ([]uint64, error) {
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

func (coll *Collection) RemoveEntry(id uint64) error {
	if uint32(id&0xffffffff) != coll.id {
		return ErrIDMismatch
	}

	entryID := uint32(id >> 32)
	delete(coll.entryCache, entryID)
	delete(coll.proto.Entries, entryID)
	return nil
}

func (coll *Collection) String() string {
	output := fmt.Sprintf("Collection: %s, Type: %s, ID: %v\n", coll.Name, coll.entryTypeName, coll.id)
	for _, entryID := range coll.EntryIDs(true) {
		output += fmt.Sprintf("ID:%v %s\n", entryID, coll.GetEntry(entryID))
	}
	return output
}

func (coll *Collection) unmarshal(bytes []byte) error {
	return coll.proto.Unmarshal(bytes)
}

func (coll *Collection) marshal() ([]byte, error) {
	for id, msg := range coll.entryCache {
		var err error
		coll.proto.Entries[id], err = msg.Marshal()
		if err != nil {
			return nil, err
		}
	}
	coll.entryCache = make(map[uint32]Entry)

	return coll.proto.Marshal()
}

func (coll *Collection) newID() uint32 {
	coll.proto.NUniqueEntryIDs++
	return coll.proto.NUniqueEntryIDs
}

func (ids idSlice) Len() int {
	return len(ids)
}

func (ids idSlice) Swap(i, j int) {
	ids[i], ids[j] = ids[j], ids[i]
}

func (ids idSlice) Less(i, j int) bool {
	return ids[i] < ids[j]
}
