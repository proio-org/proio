package eicio

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
)

type Event struct {
	Header  *EventHeader
	payload []byte
}

func NewEvent() *Event {
	return &Event{}
}

func (evt *Event) String() string {
	buffer := &bytes.Buffer{}
	fmt.Fprint(buffer, "Header: ", *evt.Header)
	fmt.Fprint(buffer, "\n\tTotal payload size: ", len(evt.payload))
	return string(buffer.Bytes())
}

func (evt *Event) AddCollection(collection proto.Message, name string) {
	collHdr := &EventHeader_CollectionHeader{}

	switch collection.(type) {
	case *MCParticleCollection:
		collHdr.Type = EventHeader_CollectionHeader_MCParticle
	case *SimTrackerHitCollection:
		collHdr.Type = EventHeader_CollectionHeader_SimTrackerHit
	}

	collHdr.Name = name

	collBuf, err := proto.Marshal(collection)
	if err != nil {
		return
	}
	collHdr.PayloadSize = uint32(len(collBuf))

	if evt.Header == nil {
		evt.Header = &EventHeader{}
	}
	evt.Header.Collections = append(evt.Header.Collections, collHdr)
	evt.payload = append(evt.payload, collBuf...)
}

func (evt *Event) GetCollection(name string) proto.Message {
	offset := uint32(0)
	size := uint32(0)
	collType := EventHeader_CollectionHeader_NONE
	for _, coll := range evt.Header.Collections {
		if coll.Name == name {
			collType = coll.Type
			size = coll.PayloadSize
			break
		}
		offset += coll.PayloadSize
	}
	if collType == EventHeader_CollectionHeader_NONE {
		return nil
	}

	var coll proto.Message
	switch collType {
	case EventHeader_CollectionHeader_MCParticle:
		coll = &MCParticleCollection{}
	case EventHeader_CollectionHeader_SimTrackerHit:
		coll = &SimTrackerHitCollection{}
	}

	if err := proto.Unmarshal(evt.payload[offset:offset+size], coll); err != nil {
		panic(err)
	}

	return coll
}

func (evt *Event) getPayload() []byte {
	return evt.payload
}

func (evt *Event) setPayload(payload []byte) {
	evt.payload = payload
}
