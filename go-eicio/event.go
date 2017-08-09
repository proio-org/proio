package eicio

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
)

type Event struct {
	Header  *EventHeader
	payload []byte
	iID     int32
}

func NewEvent() *Event {
	return &Event{Header: &EventHeader{}}
}

func (evt *Event) String() string {
	buffer := &bytes.Buffer{}

	fmt.Fprint(buffer, "Event header...\n", proto.MarshalTextString(evt.Header), "\n")
	for _, collHdr := range evt.Header.Collection {
		coll := evt.GetCollection(collHdr.Name)
		fmt.Fprint(buffer, collHdr.Name, " collection\n", proto.MarshalTextString(coll), "\n")
	}

	return string(buffer.Bytes())
}

func (evt *Event) AddCollection(collection Message, name string) {
	collHdr := &EventHeader_CollectionHeader{}

	switch collection.(type) {
	case *MCParticleCollection:
		collHdr.Type = EventHeader_CollectionHeader_MCParticle
	case *SimTrackerHitCollection:
		collHdr.Type = EventHeader_CollectionHeader_SimTrackerHit
	case *TrackerRawDataCollection:
		collHdr.Type = EventHeader_CollectionHeader_TrackerRawData
	case *TrackerDataCollection:
		collHdr.Type = EventHeader_CollectionHeader_TrackerData
	case *TrackerHitCollection:
		collHdr.Type = EventHeader_CollectionHeader_TrackerHit
	case *TrackerPulseCollection:
		collHdr.Type = EventHeader_CollectionHeader_TrackerPulse
	case *TrackerHitPlaneCollection:
		collHdr.Type = EventHeader_CollectionHeader_TrackerHitPlane
	case *TrackerHitZCylinderCollection:
		collHdr.Type = EventHeader_CollectionHeader_TrackerHitZCylinder
	case *TrackCollection:
		collHdr.Type = EventHeader_CollectionHeader_Track
	case *SimCalorimeterHitCollection:
		collHdr.Type = EventHeader_CollectionHeader_SimCalorimeterHit
	case *RawCalorimeterHitCollection:
		collHdr.Type = EventHeader_CollectionHeader_RawCalorimeterHit
	case *CalorimeterHitCollection:
		collHdr.Type = EventHeader_CollectionHeader_CalorimeterHit
	case *ClusterCollection:
		collHdr.Type = EventHeader_CollectionHeader_Cluster
	case *RecParticleCollection:
		collHdr.Type = EventHeader_CollectionHeader_RecParticle
	case *VertexCollection:
		collHdr.Type = EventHeader_CollectionHeader_Vertex
	}

	collHdr.Name = name

	collBuf, err := collection.Marshal()
	if err != nil {
		return
	}
	collHdr.PayloadSize = uint32(len(collBuf))

	if evt.Header == nil {
		evt.Header = &EventHeader{}
	}
	evt.Header.Collection = append(evt.Header.Collection, collHdr)
	evt.payload = append(evt.payload, collBuf...)
}

func (evt *Event) GetCollection(name string) Message {
	offset := uint32(0)
	size := uint32(0)
	collType := EventHeader_CollectionHeader_NONE
	for _, coll := range evt.Header.Collection {
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

	var coll Message
	switch collType {
	case EventHeader_CollectionHeader_MCParticle:
		coll = &MCParticleCollection{}
	case EventHeader_CollectionHeader_SimTrackerHit:
		coll = &SimTrackerHitCollection{}
	case EventHeader_CollectionHeader_TrackerRawData:
		coll = &TrackerRawDataCollection{}
	case EventHeader_CollectionHeader_TrackerData:
		coll = &TrackerDataCollection{}
	case EventHeader_CollectionHeader_TrackerHit:
		coll = &TrackerHitCollection{}
	case EventHeader_CollectionHeader_TrackerPulse:
		coll = &TrackerPulseCollection{}
	case EventHeader_CollectionHeader_TrackerHitPlane:
		coll = &TrackerHitPlaneCollection{}
	case EventHeader_CollectionHeader_TrackerHitZCylinder:
		coll = &TrackerHitZCylinderCollection{}
	case EventHeader_CollectionHeader_Track:
		coll = &TrackCollection{}
	case EventHeader_CollectionHeader_SimCalorimeterHit:
		coll = &SimCalorimeterHitCollection{}
	case EventHeader_CollectionHeader_RawCalorimeterHit:
		coll = &RawCalorimeterHitCollection{}
	case EventHeader_CollectionHeader_CalorimeterHit:
		coll = &CalorimeterHitCollection{}
	case EventHeader_CollectionHeader_Cluster:
		coll = &ClusterCollection{}
	case EventHeader_CollectionHeader_RecParticle:
		coll = &RecParticleCollection{}
	case EventHeader_CollectionHeader_Vertex:
		coll = &VertexCollection{}
	}

	if err := coll.Unmarshal(evt.payload[offset : offset+size]); err != nil {
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

type Message interface {
	proto.Message

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Identifiable interface {
	Message

	GetId() uint32
}
