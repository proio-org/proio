package proio

import (
	"bytes"
	"errors"
	"testing"

	"github.com/decibelcooper/proio/go-proio/model/eic"
	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
	protobuf "github.com/golang/protobuf/proto"
)

func TestStrip1(t *testing.T) {
	event := NewEvent()
	for i := 0; i < 100; i++ {
		event.AddEntry(
			"Particle",
			&eic.Particle{},
		)
	}

	for _, ID := range event.TaggedEntries("Particle") {
		event.RemoveEntry(ID)
	}

	nEntries := len(event.AllEntries())
	if nEntries > 0 {
		t.Errorf("There should be no entries, but len(event.AllEntries()) = %v out of 100", nEntries)
	}
}

func TestTagUntag1(t *testing.T) {
	event := NewEvent()
	id0 := event.AddEntry(
		"MCParticles",
		&prolcio.MCParticle{},
	)
	id1 := event.AddEntry(
		"MCParticles",
		&prolcio.MCParticle{},
	)
	event.UntagEntry(id0, "MCParticles")

	mcIDs := event.TaggedEntries("MCParticles")
	if len(mcIDs) != 1 {
		t.Errorf("%v IDs instead of 1", len(mcIDs))
	}
	if mcIDs[0] != id1 {
		t.Errorf("got ID %v instead of %v", mcIDs[0], id1)
	}
}

func TestTagUntag2(t *testing.T) {
	event := NewEvent()
	event.AddEntry(
		"MCParticles",
		&prolcio.MCParticle{},
	)
	event.AddEntry(
		"MCParticles",
		&prolcio.MCParticle{},
	)
	event.DeleteTag("MCParticles")

	mcIDs := event.TaggedEntries("MCParticles")
	if len(mcIDs) != 0 {
		t.Errorf("%v IDs instead of 0", len(mcIDs))
	}
}

func TestRevTagLookup(t *testing.T) {
	event := NewEvent()
	id := event.AddEntry(
		"MCParticles",
		&prolcio.MCParticle{},
	)
	event.TagEntry(id, "Simulated", "Particles")

	tags := event.EntryTags(id)
	if tags[0] != "MCParticles" {
		t.Errorf("First tag is %v instead of MCParticles", tags[0])
	}
	if tags[1] != "Particles" {
		t.Errorf("Second tag is %v instead of Particles", tags[1])
	}
	if tags[2] != "Simulated" {
		t.Errorf("Third tag is %v instead of Simulated", tags[2])
	}
}

func TestGetTags(t *testing.T) {
	event := NewEvent()
	id := event.AddEntry(
		"MCParticles",
		&prolcio.MCParticle{},
	)
	event.TagEntry(id, "Simulated", "Particles")

	tags := event.Tags()
	if tags[0] != "MCParticles" {
		t.Errorf("First tag is %v instead of MCParticles", tags[0])
	}
	if tags[1] != "Particles" {
		t.Errorf("Second tag is %v instead of Particles", tags[1])
	}
	if tags[2] != "Simulated" {
		t.Errorf("Third tag is %v instead of Simulated", tags[2])
	}
}

func TestDeleteTag(t *testing.T) {
	event := NewEvent()
	id := event.AddEntry(
		"MCParticles",
		&prolcio.MCParticle{},
	)
	event.TagEntry(id, "Simulated", "Particles")

	event.DeleteTag("Particles")

	tags := event.Tags()
	if tags[0] != "MCParticles" {
		t.Errorf("First tag is %v instead of MCParticles", tags[0])
	}
	if tags[1] != "Simulated" {
		t.Errorf("Second tag is %v instead of Simulated", tags[2])
	}
}

func TestDirtyTag(t *testing.T) {
	event := NewEvent()
	id0 := event.AddEntry(
		"Particle",
		&eic.Particle{},
	)
	id1 := event.AddEntry(
		"Particle",
		&eic.Particle{},
	)
	event.RemoveEntry(id0)

	ids := event.TaggedEntries("Particle")
	if len(ids) != 1 {
		t.Errorf("%v IDs instead of 1", len(ids))
	}
	if ids[0] != id1 {
		t.Errorf("got ID %v instead of %v", ids[0], id1)
	}
}

func TestNoSuchEntry(t *testing.T) {
	event := NewEvent()
	entry := event.GetEntry(0)
	if entry != nil {
		t.Error("Entry is not nil")
	}
}

type unknownMsg struct {
}

func (*unknownMsg) Reset()         {}
func (*unknownMsg) String() string { return "" }
func (*unknownMsg) ProtoMessage()  {}

func TestUnknownType(t *testing.T) {
	event := NewEvent()
	id := event.AddEntry("unknown", &unknownMsg{})

	writer := NewWriter(&bytes.Buffer{})
	writer.Push(event)

	entry := event.GetEntry(id)
	if entry != nil {
		t.Error("Event returns entry for unknown type")
	}
}

type nonSelfSerializingMsg struct {
	X float32 `protobuf:"fixed32,1,opt,name=x,proto3" json:"x,omitempty"`
	Y float32 `protobuf:"fixed32,2,opt,name=y,proto3" json:"y,omitempty"`
	Z float32 `protobuf:"fixed32,3,opt,name=z,proto3" json:"z,omitempty"`
}

func (*nonSelfSerializingMsg) Reset()         {}
func (*nonSelfSerializingMsg) String() string { return "" }
func (*nonSelfSerializingMsg) ProtoMessage()  {}

func init() {
	protobuf.RegisterType((*nonSelfSerializingMsg)(nil), "nonSelfSerializingMsg")
}

func TestNonSelfSerializingMsg(t *testing.T) {
	event := NewEvent()
	id := event.AddEntry("nonSelfSerializingMsg", &nonSelfSerializingMsg{})

	writer := NewWriter(&bytes.Buffer{})
	writer.Push(event)

	entry := event.GetEntry(id)
	if entry == nil {
		t.Error("Unable to deserialize message")
	}
}

type halfSelfSerializingMsg1 struct {
	X float32 `protobuf:"fixed32,1,opt,name=x,proto3" json:"x,omitempty"`
	Y float32 `protobuf:"fixed32,2,opt,name=y,proto3" json:"y,omitempty"`
	Z float32 `protobuf:"fixed32,3,opt,name=z,proto3" json:"z,omitempty"`
}

func (*halfSelfSerializingMsg1) Reset()                   {}
func (*halfSelfSerializingMsg1) String() string           { return "" }
func (*halfSelfSerializingMsg1) ProtoMessage()            {}
func (*halfSelfSerializingMsg1) Marshal() ([]byte, error) { return []byte{0x7}, nil }

func init() {
	protobuf.RegisterType((*halfSelfSerializingMsg1)(nil), "halfSelfSerializingMsg1")
}

func TestHalfSelfSerializingMsg1(t *testing.T) {
	event := NewEvent()
	id := event.AddEntry("halfSelfSerializingMsg1", &halfSelfSerializingMsg1{})

	writer := NewWriter(&bytes.Buffer{})
	writer.Push(event)

	entry := event.GetEntry(id)
	if entry != nil {
		t.Error("Broken message returns non-nil value")
	}
}

type halfSelfSerializingMsg2 struct {
	X float32 `protobuf:"fixed32,1,opt,name=x,proto3" json:"x,omitempty"`
	Y float32 `protobuf:"fixed32,2,opt,name=y,proto3" json:"y,omitempty"`
	Z float32 `protobuf:"fixed32,3,opt,name=z,proto3" json:"z,omitempty"`
}

func (*halfSelfSerializingMsg2) Reset()                 {}
func (*halfSelfSerializingMsg2) String() string         { return "" }
func (*halfSelfSerializingMsg2) ProtoMessage()          {}
func (*halfSelfSerializingMsg2) Unmarshal([]byte) error { return errors.New("bad") }

func init() {
	protobuf.RegisterType((*halfSelfSerializingMsg2)(nil), "halfSelfSerializingMsg2")
}

func TestHalfSelfSerializingMsg2(t *testing.T) {
	event := NewEvent()
	id := event.AddEntry("halfSelfSerializingMsg2", &halfSelfSerializingMsg2{})

	writer := NewWriter(&bytes.Buffer{})
	writer.Push(event)

	entry := event.GetEntry(id)
	if entry != nil {
		t.Error("Broken message returns non-nil value")
	}
}
