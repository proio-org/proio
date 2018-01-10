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
	Err error

	proto *proto.Event

	revTypeLookup  map[string]uint64
	revTagLookup   map[uint64][]string
	entryTypeCache map[uint64]reflect.Type
	entryCache     map[uint64]protobuf.Message
}

// NewEvent is required for constructing an Event.
func NewEvent() *Event {
	return &Event{
		proto: &proto.Event{
			Entries: make(map[uint64]*proto.Any),
			Types:   make(map[uint64]string),
			Tags:    make(map[string]*proto.Tag),
		},
		revTypeLookup:  make(map[string]uint64),
		revTagLookup:   make(map[uint64][]string),
		entryTypeCache: make(map[uint64]reflect.Type),
		entryCache:     make(map[uint64]protobuf.Message),
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
	var ids []uint64
	for _, entry := range entries {
		ids = append(ids, evt.AddEntry(tag, entry))
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
	selfSerializingEntry, ok := entry.(selfSerializingEntry)
	if ok {
		if err := selfSerializingEntry.Unmarshal(entryProto.Payload); err != nil {
			evt.Err = errors.New(
				"failure to unmarshal entry " +
					strconv.FormatUint(id, 10) +
					" with type " +
					evt.proto.Types[entryProto.Type],
			)
			return nil
		}
	} else {
		if err := protobuf.Unmarshal(entryProto.Payload, entry); err != nil {
			evt.Err = errors.New(
				"failure to unmarshal entry " +
					strconv.FormatUint(id, 10) +
					" with type " +
					evt.proto.Types[entryProto.Type],
			)
			return nil
		}
	}

	evt.entryCache[id] = entry

	evt.Err = nil
	return entry
}

func (evt *Event) RemoveEntry(id uint64) {
	tags := evt.EntryTags(id)
	for _, tag := range tags {
		tagProto := evt.proto.Tags[tag]
		for i, thisID := range tagProto.Entries {
			if thisID == id {
				tagProto.Entries = append(tagProto.Entries[:i], tagProto.Entries[i+1:]...)
			}
		}
	}

	delete(evt.revTagLookup, id)
	delete(evt.entryCache, id)
	delete(evt.proto.Entries, id)
}

func (evt *Event) AllEntries() []uint64 {
	var IDs []uint64
	for ID, _ := range evt.proto.Entries {
		IDs = append(IDs, ID)
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
		return tagProto.Entries[:]
	}
	return nil
}

// Tags returns a list of all tags in the Event.
func (evt *Event) Tags() []string {
	var tags []string
	for key, _ := range evt.proto.Tags {
		tags = append(tags, key)
	}
	sort.Strings(tags)
	return tags
}

// EntryTags does a reverse lookup of tags that point to a given entry ID.
func (evt *Event) EntryTags(id uint64) []string {
	tags, ok := evt.revTagLookup[id]
	if ok {
		return tags
	}

	tags = make([]string, 0)
	for name, tagProto := range evt.proto.Tags {
		for _, thisID := range tagProto.Entries {
			if thisID == id {
				tags = append(tags, name)
				break
			}
		}
	}
	sort.Strings(tags)

	evt.revTagLookup[id] = tags

	return tags
}

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

type selfSerializingEntry interface {
	protobuf.Message

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
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
		proto:          eventProto,
		revTypeLookup:  make(map[string]uint64),
		revTagLookup:   make(map[uint64][]string),
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
		selfSerializingEntry, ok := entry.(selfSerializingEntry)
		var bytes []byte
		if ok {
			bytes, _ = selfSerializingEntry.Marshal()
		} else {
			bytes, _ = protobuf.Marshal(entry)
		}
		evt.proto.Entries[id].Payload = bytes
	}
	evt.entryCache = make(map[uint64]protobuf.Message)
}

func fromProto(bytes []byte) *Event {
	eventProto := &proto.Event{}
	err := eventProto.Unmarshal(bytes)
	if err != nil {
		return nil
	}
	return &Event{
		proto: eventProto,
	}
}
