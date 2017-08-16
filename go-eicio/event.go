package eicio

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
)

type Event struct {
	Header  *EventHeader
	payload []byte

	collCache   map[string]Collection
	namesCached []string
}

func NewEvent() *Event {
	return &Event{
		Header:    &EventHeader{},
		collCache: make(map[string]Collection),
	}
}

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

func GetType(coll Collection) string {
	return strings.TrimPrefix(proto.MessageName(coll), "eicio.")
}

var (
	ErrDupCollection = errors.New("duplicate collection name")
	ErrDupID         = errors.New("duplicate collection id")
)

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
// removed from Header.PayloadCollection
func (evt *Event) Get(name string) (Collection, error) {
	if msg := evt.collCache[name]; msg != nil {
		return msg, nil
	}

	return evt.getFromPayload(name, true)
}

var ErrMsgNotFound = errors.New("unable to reference: message not found")

func (evt *Event) GetUniqueID() uint32 {
	evt.Header.NUniqueIDs++
	return evt.Header.NUniqueIDs
}

func (evt *Event) Reference(msg Message) (*Reference, error) {
	for _, coll := range evt.collCache {
		if coll == msg {
			collID := coll.GetId()
			if collID == 0 {
				collID = evt.GetUniqueID()
				coll.SetId(collID)
			}
			return &Reference{
				CollID:  collID,
				EntryID: 0,
			}, nil
		}

		for i := uint32(0); i < coll.GetNEntries(); i++ {
			entry := coll.GetEntry(i)
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
				return &Reference{
					CollID:  collID,
					EntryID: entryID,
				}, nil
			}
		}
	}

	return nil, ErrMsgNotFound
}

func (evt *Event) Dereference(ref *Reference) (Message, error) {
	var refColl Collection
	for _, coll := range evt.collCache {
		if coll.GetId() == ref.CollID {
			if ref.EntryID == 0 {
				return coll, nil
			}
			refColl = coll
			break
		}
	}
	if refColl == nil {
		payloadCollections := make([]*EventHeader_CollectionHeader, len(evt.Header.PayloadCollections))
		copy(evt.Header.PayloadCollections, payloadCollections)
		for _, collHdr := range payloadCollections {
			if collHdr.Id == ref.CollID {
				var err error
				if refColl, err = evt.Get(collHdr.Name); err != nil {
					return nil, err
				}
				break
			}
		}
	}
	if refColl == nil {
		return nil, ErrMsgNotFound
	}

	for i := uint32(0); i < refColl.GetNEntries(); i++ {
		entry := refColl.GetEntry(i)
		if entry.GetId() == ref.EntryID {
			return entry, nil
		}
	}
	return nil, ErrMsgNotFound
}

func (evt *Event) String() string {
	buffer := &bytes.Buffer{}

	stringBuf := fmt.Sprint(evt.Header, "\n")
	stringBuf = strings.Replace(stringBuf, " payloadCollections:", "\n\tpayloadCollections:", -1)
	stringBuf = strings.Replace(stringBuf, " >", ">", -1)
	fmt.Fprint(buffer, stringBuf, "\n")

	for _, name := range evt.GetNames() {
		coll, _ := evt.Get(name)
		if coll != nil {
			fmt.Fprint(buffer, "\tname:", name, " type:", GetType(coll), "\n")

			stringBuf = fmt.Sprint("\t\t", coll, "\n")
			stringBuf = strings.Replace(stringBuf, " entries:", "\n\t\tentries:", -1)
			stringBuf = strings.Replace(stringBuf, " >", ">", -1)
			fmt.Fprint(buffer, stringBuf)
		}
	}

	return string(buffer.Bytes())
}

var ErrBlankColl = errors.New("collection not found or type is blank")

func (evt *Event) getFromPayload(name string, unmarshal bool) (Collection, error) {
	offset := uint32(0)
	size := uint32(0)
	collType := ""
	collIndex := 0
	var collHdr *EventHeader_CollectionHeader
	for collIndex, collHdr = range evt.Header.PayloadCollections {
		if collHdr.Name == name {
			collType = collHdr.Type
			size = collHdr.PayloadSize
			break
		}
		offset += collHdr.PayloadSize
	}
	if collType == "" {
		return nil, ErrBlankColl
	}

	var coll Collection
	if unmarshal {
		msgType := proto.MessageType("eicio." + collType).Elem()
		coll = reflect.New(msgType).Interface().(Collection)
		if err := coll.Unmarshal(evt.payload[offset : offset+size]); err != nil {
			return nil, err
		}

		evt.collCache[name] = coll
	}

	evt.namesCached = append(evt.namesCached, name)
	evt.Header.PayloadCollections = append(evt.Header.PayloadCollections[:collIndex], evt.Header.PayloadCollections[collIndex+1:]...)
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

func (evt *Event) collToPayload(coll Collection, name string) error {
	collHdr := &EventHeader_CollectionHeader{}
	collHdr.Name = name
	collHdr.Id = coll.GetId()
	collHdr.Type = GetType(coll)

	collBuf, err := coll.Marshal()
	if err != nil {
		return err
	}
	collHdr.PayloadSize = uint32(len(collBuf))

	if evt.Header == nil {
		evt.Header = &EventHeader{}
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
	GetEntry(uint32) Message
}
