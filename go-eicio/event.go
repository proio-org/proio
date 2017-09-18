package eicio // import "github.com/decibelcooper/eicio/go-eicio"

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/decibelcooper/eicio/go-eicio/model"
)

// Struct representing an event either created with NewEvent() or retrieved
// with (*Reader) Next()
type Event struct {
	Header  *model.EventHeader
	payload []byte

	collCache   map[string]Collection
	namesCached []string
}

// Returns a new event with minimal initialization
func NewEvent() *Event {
	return &Event{
		Header:    &model.EventHeader{},
		collCache: make(map[string]Collection),
	}
}

// Get a list of collection names in the event
func (evt *Event) GetNames() []string {
	names := make([]string, 0)

	for _, collHdr := range evt.Header.PayloadCollections {
		names = append(names, collHdr.Name)
	}
	for name, _ := range evt.collCache {
		names = append(names, name)
	}

	return names
}

// Get a string representing the collection type
func GetType(coll Collection) string {
	return strings.TrimPrefix(proto.MessageName(coll), "eicio.model.")
}

var (
	ErrDupCollection = errors.New("duplicate collection name")
	ErrDupID         = errors.New("duplicate collection id")
)

// Add a collection to the event.  This allows automatic referencing with
// (*Event) Reference(), and queues the collection for serialization once
// (*Writer) Push() is called.  The collection may be modified further after
// adding.
func (evt *Event) Add(coll Collection, name string) error {
	for key, coll_ := range evt.collCache {
		if key == name {
			return ErrDupCollection
		}
		if coll_.GetId() != 0 && coll_.GetId() == coll.GetId() {
			return ErrDupID
		}
	}

	for _, collHdr := range evt.Header.PayloadCollections {
		if collHdr.Name == name {
			return ErrDupCollection
		}
		if collHdr.Id != 0 && collHdr.Id == coll.GetId() {
			return ErrDupID
		}
	}

	evt.collCache[name] = coll
	evt.namesCached = append(evt.namesCached, name)
	return nil
}

// Remove a collection from the event by name
func (evt *Event) Remove(name string) {
	for key, _ := range evt.collCache {
		if key == name {
			delete(evt.collCache, key)
			return
		}
	}
	for _, collHdr := range evt.Header.PayloadCollections {
		if collHdr.Name == name {
			evt.getFromPayload(name, false)
			return
		}
	}
}

// Gets a collection from the event.  The collection is deserialized upon the
// first time calling this function.  Once deserialized, the collection is
// removed from Header.PayloadCollection, and placed back into a queue for
// reserialization.  The event may be safely modified before reserializing.
func (evt *Event) Get(name string) Collection {
	if msg := evt.collCache[name]; msg != nil {
		return msg
	}

	return evt.getFromPayload(name, true)
}

// Get a unique ID for referencing collections or entries.  This is typically
// not needed, and instead (*Event) Reference() should be called.
func (evt *Event) GetUniqueID() uint32 {
	evt.Header.NUniqueIDs++
	return evt.Header.NUniqueIDs
}

// Reference a message that exists in the event.  The message may be a
// collection or entry within a collection.  In both cases, the collection must
// have been added to the event, or it must be a collection retrieved with
// (*Event) Get().
func (evt *Event) Reference(msg Message) *model.Reference {
	for _, coll := range evt.collCache {
		if coll == msg {
			collID := coll.GetId()
			if collID == 0 {
				collID = evt.GetUniqueID()
				coll.SetId(collID)
			}
			return &model.Reference{
				CollID:  collID,
				EntryID: 0,
			}
		}

		for i := uint32(0); i < coll.GetNEntries(); i++ {
			entry := coll.GetEntry(i).(Message)
			if entry == msg {
				collID := coll.GetId()
				if collID == 0 {
					collID = evt.GetUniqueID()
					coll.SetId(collID)
				}
				entryID := entry.GetId()
				if entryID == 0 {
					entryID = evt.GetUniqueID()
					entry.SetId(entryID)
				}
				return &model.Reference{
					CollID:  collID,
					EntryID: entryID,
				}
			}
		}
	}

	return nil
}

// Dereference a message from the event.  This returns a message (either
// collection or collection entry) referred to by a Reference.  The message
// must exist in the event.
func (evt *Event) Dereference(ref *model.Reference) Message {
	var refColl Collection
	for _, coll := range evt.collCache {
		if coll.GetId() == ref.CollID {
			if ref.EntryID == 0 {
				return coll
			}
			refColl = coll
			break
		}
	}
	if refColl == nil {
		for _, collHdr := range evt.Header.PayloadCollections {
			if collHdr.Id == ref.CollID {
				if refColl = evt.Get(collHdr.Name); refColl == nil {
					return nil
				}
				break
			}
		}
	}
	if refColl == nil {
		return nil
	}

	for i := uint32(0); i < refColl.GetNEntries(); i++ {
		entry := refColl.GetEntry(i).(Message)
		if entry.GetId() == ref.EntryID {
			return entry
		}
	}
	return nil
}

func (evt *Event) String() string {
	buffer := &bytes.Buffer{}

	stringBuf := fmt.Sprint(evt.Header, "\n")
	stringBuf = strings.Replace(stringBuf, " payloadCollections:", "\n    payloadCollections:", -1)
	stringBuf = strings.Replace(stringBuf, " >", ">", -1)
	fmt.Fprint(buffer, stringBuf, "\n")

	for _, name := range evt.GetNames() {
		coll := evt.Get(name)
		if coll != nil {
			fmt.Fprint(buffer, "    name:", name, " type:", GetType(coll), "\n")

			stringBuf = fmt.Sprint("        ", coll, "\n")
			stringBuf = strings.Replace(stringBuf, " entries:", "\n        entries:", -1)
			stringBuf = strings.Replace(stringBuf, " >", ">", -1)
			fmt.Fprint(buffer, stringBuf)
		}
	}

	return string(buffer.Bytes())
}

func (evt *Event) getFromPayload(name string, unmarshal bool) Collection {
	offset := uint32(0)
	size := uint32(0)
	collType := ""
	collIndex := 0
	var collHdr *model.EventHeader_CollectionHeader
	for collIndex, collHdr = range evt.Header.PayloadCollections {
		if collHdr.Name == name {
			collType = collHdr.Type
			size = collHdr.PayloadSize
			break
		}
		offset += collHdr.PayloadSize
	}
	if collType == "" {
		return nil
	}

	var coll Collection
	if unmarshal {
		msgType := proto.MessageType("eicio.model." + collType).Elem()
		coll = reflect.New(msgType).Interface().(Collection)
		if err := coll.Unmarshal(evt.payload[offset : offset+size]); err != nil {
			return nil
		}

		evt.collCache[name] = coll
	}

	evt.namesCached = append(evt.namesCached, name)
	evt.Header.PayloadCollections = append(evt.Header.PayloadCollections[:collIndex], evt.Header.PayloadCollections[collIndex+1:]...)
	evt.payload = append(evt.payload[:offset], evt.payload[offset+size:]...)

	return coll
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

func (evt *Event) collToPayload(coll Collection, name string) error {
	collHdr := &model.EventHeader_CollectionHeader{}
	collHdr.Name = name
	collHdr.Id = coll.GetId()
	collHdr.Type = GetType(coll)

	collBuf, err := coll.Marshal()
	if err != nil {
		return err
	}
	collHdr.PayloadSize = uint32(len(collBuf))

	if evt.Header == nil {
		evt.Header = &model.EventHeader{}
	}
	evt.Header.PayloadCollections = append(evt.Header.PayloadCollections, collHdr)
	evt.payload = append(evt.payload, collBuf...)

	return nil
}

func (evt *Event) getPayload() []byte {
	return evt.payload
}

func (evt *Event) setPayload(payload []byte) {
	evt.payload = payload
}

type Message interface {
	proto.Message

	Marshal() ([]byte, error)
	Unmarshal([]byte) error

	GetId() uint32
	SetId(uint32)
}

type Collection interface {
	Message

	GetNEntries() uint32
	GetEntry(uint32) proto.Message
}
