package proio

import (
	"bytes"
	"testing"

	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

func TestRefDeref1(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	parent := &prolcio.MCParticle{PDG: 443}
	parentID := eventOut.AddEntry("MCParticles", parent)
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: -11}
	childIDs := eventOut.AddEntries("MCParticles", child1, child2)
	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	writer.Push(eventOut)
	writer.Flush()

	reader := NewReader(buffer)

	eventIn, err := reader.Next()
	if err != nil {
		t.Error("Error reading back event: ", err)
	}

	MCParticles := eventIn.TaggedEntries("MCParticles")
	if MCParticles == nil {
		t.Error("Failed to get MCParticles tag")
	}

	parent_ := eventIn.GetEntry(MCParticles[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match first prolcio.MCParticle")
	}
	child1_ := eventIn.GetEntry(parent_.Children[0]).(*prolcio.MCParticle)
	if child1_.String() != child1.String() {
		t.Error("Failed to match first daughter particle")
	}
	child2_ := eventIn.GetEntry(parent_.Children[1]).(*prolcio.MCParticle)
	if child1_.String() != child1.String() {
		t.Error("Failed to match second daughter particle")
	}
	parent_ = eventIn.GetEntry(child1_.Parents[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match parent of first daughter particle")
	}
	parent_ = eventIn.GetEntry(child2_.Parents[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match parent of second daughter particle")
	}
}

func TestRefDeref2(t *testing.T) {
	event := NewEvent()
	parent := &prolcio.MCParticle{PDG: 443}
	parentID := event.AddEntry("MCParticles", parent)
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: -11}
	childIDs := event.AddEntries("MCParticles", child1, child2)
	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	MCParticles := event.TaggedEntries("MCParticles")
	if MCParticles == nil {
		t.Error("Failed to get MCParticles tag")
	}

	parent_ := event.GetEntry(MCParticles[0]).(*prolcio.MCParticle)
	if parent_ != parent {
		t.Error("Failed to match first prolcio.MCParticle")
	}
	child1_ := event.GetEntry(parent_.Children[0]).(*prolcio.MCParticle)
	if child1_ != child1 {
		t.Error("Failed to match first daughter particle")
	}
	child2_ := event.GetEntry(parent_.Children[1]).(*prolcio.MCParticle)
	if child1_ != child1 {
		t.Error("Failed to match second daughter particle")
	}
	parent_ = event.GetEntry(child1_.Parents[0]).(*prolcio.MCParticle)
	if parent_ != parent {
		t.Error("Failed to match parent of first daughter particle")
	}
	parent_ = event.GetEntry(child2_.Parents[0]).(*prolcio.MCParticle)
	if parent_ != parent {
		t.Error("Failed to match parent of second daughter particle")
	}
}
