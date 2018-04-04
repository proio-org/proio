package proio // import "github.com/decibelcooper/proio/go-proio"

// Generate protobuf messages
//go:generate bash gen.sh

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/decibelcooper/proio/go-proio/proto"
	protobuf "github.com/golang/protobuf/proto"
)

// Event contains all data for an event, and provides methods for adding and
// retrieving data.
type Event struct {
	Err      error
	Metadata map[string][]byte

	proto *proto.Event

	revTypeLookup  map[string]uint64
	entryTypeCache map[uint64]reflect.Type
	entryCache     map[uint64]protobuf.Message
	dirtyTags      bool
}

// NewEvent is required for constructing an Event.
func NewEvent() *Event {
	return &Event{
		Metadata: make(map[string][]byte),
		proto: &proto.Event{
			Entries: make(map[uint64]*proto.Any),
			Types:   make(map[uint64]string),
			Tags:    make(map[string]*proto.Tag),
		},
		revTypeLookup:  make(map[string]uint64),
		entryTypeCache: make(map[uint64]reflect.Type),
		entryCache:     make(map[uint64]protobuf.Message),
		dirtyTags:      false,
	}
}

// AddEntry takes a single primary tag for an entry and an entry protobuf
// message, and returns a new ID number for the entry.  This ID number can be
// used to persistently reference the entry.  For example, pass the ID TagEntry
// to add additional tags to the entry.
func (evt *Event) AddEntry(tag string, entry protobuf.Message) uint64 {
	typeID := evt.getTypeID(entry)
	entryProto := &proto.Any{
		Type: typeID,
	}

	evt.proto.NEntries++
	id := evt.proto.NEntries
	evt.proto.Entries[id] = entryProto

	evt.entryCache[id] = entry

	evt.TagEntry(id, tag)

	return id
}

// AddEntries is like AddEntry, except that it is variadic, taking an arbitrary
// number of entries separated by commas.  Additionally, the return value is a
// slice of IDs.
func (evt *Event) AddEntries(tag string, entries ...protobuf.Message) []uint64 {
	ids := make([]uint64, len(entries))
	for i, entry := range entries {
		ids[i] = evt.AddEntry(tag, entry)
	}
	return ids
}

// GetEntry retrieves and deserializes an entry corresponding to the given ID
// number.  The deserialized entry is returned.  The entry type must be one
// that has been linked (and therefore initialized) with the current
// executable, otherwise it is an unknown type and nil is returned.
func (evt *Event) GetEntry(id uint64) protobuf.Message {
	entry, ok := evt.entryCache[uint64(id)]
	if ok {
		evt.Err = nil
		return entry
	}

	entryProto, ok := evt.proto.Entries[uint64(id)]
	if !ok {
		evt.Err = errors.New("no such entry: " + strconv.FormatUint(id, 10))
		return nil
	}

	entry = evt.getPrototype(entryProto.Type)
	if entry == nil {
		evt.Err = errors.New("unknown type: " + evt.proto.Types[entryProto.Type])
		return nil
	}

	if err := protobuf.Unmarshal(entryProto.Payload, entry); err != nil {
		evt.Err = errors.New(
			"failure to unmarshal entry " +
				strconv.FormatUint(id, 10) +
				" with type " +
				evt.proto.Types[entryProto.Type],
		)
		return nil
	}

	evt.entryCache[id] = entry

	evt.Err = nil
	return entry
}

// RemoveEntry takes an entry id and removes the referenced entry from the
// Event.
func (evt *Event) RemoveEntry(id uint64) {
	delete(evt.entryCache, id)
	delete(evt.proto.Entries, id)
	evt.dirtyTags = true
}

// AllEntries returns a slice of identifiers for all entries contained in the
// Event.
func (evt *Event) AllEntries() []uint64 {
	IDs := make([]uint64, len(evt.proto.Entries))
	var i int
	for ID := range evt.proto.Entries {
		IDs[i] = ID
		i++
	}
	return IDs
}

// TagEntry adds additional tags to an entry ID returned by AddEntry.
func (evt *Event) TagEntry(id uint64, tags ...string) {
	for _, tag := range tags {
		tagProto, ok := evt.proto.Tags[tag]
		if !ok {
			tagProto = &proto.Tag{}
			evt.proto.Tags[tag] = tagProto
		}

		tagProto.Entries = append(tagProto.Entries, id)
	}
}

// UntagEntry removes the association between a tag and an entry.
func (evt *Event) UntagEntry(id uint64, tag string) {
	tagProto, ok := evt.proto.Tags[tag]
	if !ok {
		return
	}

	for i, entryID := range tagProto.Entries {
		if entryID == id {
			tagProto.Entries = append(tagProto.Entries[:i], tagProto.Entries[i+1:]...)
			return
		}
	}
}

// TaggedEntries returns a slice of ID numbers that are referenced by the given
// tag.
func (evt *Event) TaggedEntries(tag string) []uint64 {
	tagProto, ok := evt.proto.Tags[tag]
	if ok {
		evt.tagCleanup()
		entries := make([]uint64, len(tagProto.Entries))
		copy(entries, tagProto.Entries)
		return entries
	}
	return nil
}

// Tags returns a list of all tags in the Event.
func (evt *Event) Tags() []string {
	var tags []string
	for key := range evt.proto.Tags {
		tags = append(tags, key)
	}
	sort.Strings(tags)
	return tags
}

// EntryTags does a reverse lookup of tags that point to a given entry ID.
func (evt *Event) EntryTags(id uint64) []string {
	tags := make([]string, 0)
	for name, tagProto := range evt.proto.Tags {
		for _, thisID := range tagProto.Entries {
			if thisID == id {
				tags = append(tags, name)
				break
			}
		}
	}
	sort.Strings(tags)
	return tags
}

// DeleteTag takes a tag name as an argument and deletes that tag from the
// Event if it exists.
func (evt *Event) DeleteTag(tag string) {
	delete(evt.proto.Tags, tag)
}

func (evt *Event) String() string {
	var printString string

	tags := evt.Tags()

	for _, tag := range tags {
		printString += "---------- TAG: " + tag + " ----------\n"
		entries := evt.TaggedEntries(tag)
		for _, entryID := range entries {
			printString += fmt.Sprintf("ID: %v\n", entryID)
			entry := evt.GetEntry(entryID)
			if entry != nil {
				typeName := protobuf.MessageName(entry)
				printString += "Entry type: " + typeName + "\n"
				printString += protobuf.MarshalTextString(entry) + "\n"
			} else {
				printString += evt.Err.Error() + "\n"
			}
		}
	}

	return printString
}

func newEventFromProto(eventProto *proto.Event) *Event {
	if eventProto.Entries == nil {
		eventProto.Entries = make(map[uint64]*proto.Any)
	}
	if eventProto.Types == nil {
		eventProto.Types = make(map[uint64]string)
	}
	if eventProto.Tags == nil {
		eventProto.Tags = make(map[string]*proto.Tag)
	}
	return &Event{
		Metadata:       make(map[string][]byte),
		proto:          eventProto,
		revTypeLookup:  make(map[string]uint64),
		entryTypeCache: make(map[uint64]reflect.Type),
		entryCache:     make(map[uint64]protobuf.Message),
	}
}

func (evt *Event) getPrototype(id uint64) protobuf.Message {
	entryType, ok := evt.entryTypeCache[id]
	if !ok {
		ptrType := protobuf.MessageType(evt.proto.Types[id])
		if ptrType == nil {
			return nil
		}
		entryType = ptrType.Elem()
		evt.entryTypeCache[id] = entryType
	}

	return reflect.New(entryType).Interface().(protobuf.Message)
}

func (evt *Event) getTypeID(entry protobuf.Message) uint64 {
	typeName := protobuf.MessageName(entry)
	typeID, ok := evt.revTypeLookup[typeName]
	if !ok {
		for id, name := range evt.proto.Types {
			if name == typeName {
				evt.revTypeLookup[typeName] = id
				return id
			}
		}

		evt.proto.NTypes++
		typeID = evt.proto.NTypes
		evt.proto.Types[typeID] = typeName
		evt.revTypeLookup[typeName] = typeID
	}

	return typeID
}

func (evt *Event) flushCache() {
	for id, entry := range evt.entryCache {
		selfSerializingEntry, ok := entry.(protobuf.Marshaler)
		var bytes []byte
		if ok {
			bytes, _ = selfSerializingEntry.Marshal()
		} else {
			bytes, _ = protobuf.Marshal(entry)
		}
		evt.proto.Entries[id].Payload = bytes
	}
	evt.entryCache = make(map[uint64]protobuf.Message)

	evt.tagCleanup()
}

func (evt *Event) tagCleanup() {
	if !evt.dirtyTags {
		return
	}
	for _, tagProto := range evt.proto.Tags {
		for i := len(tagProto.Entries) - 1; i >= 0; i-- {
			if _, ok := evt.proto.Entries[tagProto.Entries[i]]; !ok {
				tagProto.Entries = append(tagProto.Entries[:i], tagProto.Entries[i+1:]...)
			}
		}
	}
	evt.dirtyTags = false
}
