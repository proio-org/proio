package eicio

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEventPushGet(t *testing.T) {
	buffer := &bytes.Buffer{}

	writer := NewWriter(buffer)

	event0Out := NewEvent()

	MCParticles := &MCParticleCollection{}
	MCParticles.Entries = append(MCParticles.Entries, &MCParticle{})
	MCParticles.Entries = append(MCParticles.Entries, &MCParticle{})
	event0Out.Add(MCParticles, "MCParticles")

	simTrackHits := &SimTrackerHitCollection{}
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	event0Out.Add(simTrackHits, "TrackerHits")

	writer.PushEvent(event0Out)

	event1Out := NewEvent()

	simTrackHits = &SimTrackerHitCollection{}
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	event1Out.Add(simTrackHits, "TrackerHits")

	writer.PushEvent(event1Out)

	reader := NewReader(buffer)

	event0In, err := reader.Next()
	if err != nil {
		t.Error(err)
	}
	if event0In == nil {
		t.Error("Event 0 failed to Get")
	}
	if !reflect.DeepEqual(event0Out, event0In) {
		t.Error("Event 0 corrupted")
	}

	event1In, err := reader.Next()
	if err != nil {
		t.Error(err)
	}
	if event1In == nil {
		t.Error("Event 1 failed to Get")
	}
	if !reflect.DeepEqual(event1Out, event1In) {
		t.Error("Event 1 corrupted")
	}
}

func TestRefDeref(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()

	MCParticles := &MCParticleCollection{}
	if err := eventOut.Add(MCParticles, "MCParticles"); err != nil {
		t.Error("Can't add MCParticles collection: ", err)
	}

	part1 := &MCParticle{PDG: 11}
	MCParticles.Entries = append(MCParticles.Entries, part1)
	part2 := &MCParticle{PDG: 11}
	MCParticles.Entries = append(MCParticles.Entries, part2)
	part3 := &MCParticle{PDG: 22}
	MCParticles.Entries = append(MCParticles.Entries, part3)

	part1.Children = append(part1.Children, eventOut.Reference(part2))
	part1.Children = append(part1.Children, eventOut.Reference(part3))
	part2.Parents = append(part2.Parents, eventOut.Reference(part1))
	part3.Parents = append(part3.Parents, eventOut.Reference(part1))

	writer.PushEvent(eventOut)

	reader := NewReader(buffer)

	eventIn, err := reader.Next()
	if err != nil {
		t.Error("Error reading back event")
	}

	MCParticles_, err := eventIn.Get("MCParticles")
	if err != nil {
		t.Error("Failed to get MCParticles collection")
	}

	part1_ := MCParticles_.GetEntry(0).(*MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match first MCParticle")
	}
	part2_ := eventOut.Dereference(part1_.Children[0]).(*MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match first daughter particle")
	}
	part3_ := eventOut.Dereference(part1_.Children[1]).(*MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match second daughter particle")
	}
	part1_ = eventOut.Dereference(part2_.Parents[0]).(*MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of first daughter particle")
	}
	part1_ = eventOut.Dereference(part3_.Parents[0]).(*MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of second daughter particle")
	}
}

func TestRefDeref2(t *testing.T) {
	event := NewEvent()

	MCParticles := &MCParticleCollection{}
	if err := event.Add(MCParticles, "MCParticles"); err != nil {
		t.Error("Can't add MCParticles collection: ", err)
	}

	part1 := &MCParticle{PDG: 11}
	MCParticles.Entries = append(MCParticles.Entries, part1)
	part2 := &MCParticle{PDG: 11}
	MCParticles.Entries = append(MCParticles.Entries, part2)
	part3 := &MCParticle{PDG: 22}
	MCParticles.Entries = append(MCParticles.Entries, part3)

	part1.Children = append(part1.Children, event.Reference(part2))
	part1.Children = append(part1.Children, event.Reference(part3))
	part2.Parents = append(part2.Parents, event.Reference(part1))
	part3.Parents = append(part3.Parents, event.Reference(part1))

	MCParticles_, err := event.Get("MCParticles")
	if err != nil {
		t.Error("Failed to get MCParticles collection")
	}

	part1_ := MCParticles_.GetEntry(0).(*MCParticle)
	if part1_ != part1 {
		t.Error("Failed to match first MCParticle")
	}
	part2_ := event.Dereference(part1_.Children[0]).(*MCParticle)
	if part2_ != part2 {
		t.Error("Failed to match first daughter particle")
	}
	part3_ := event.Dereference(part1_.Children[1]).(*MCParticle)
	if part2_ != part2 {
		t.Error("Failed to match second daughter particle")
	}
	part1_ = event.Dereference(part2_.Parents[0]).(*MCParticle)
	if part1_ != part1 {
		t.Error("Failed to match parent of first daughter particle")
	}
	part1_ = event.Dereference(part3_.Parents[0]).(*MCParticle)
	if part1_ != part1 {
		t.Error("Failed to match parent of second daughter particle")
	}
}

func TestRefDeref3(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()

	MCParticles := &MCParticleCollection{}
	if err := eventOut.Add(MCParticles, "MCParticles"); err != nil {
		t.Error("Can't add MCParticles collection: ", err)
	}
	part1 := &MCParticle{PDG: 11}
	MCParticles.Entries = append(MCParticles.Entries, part1)

	SimParticles := &MCParticleCollection{}
	if err := eventOut.Add(SimParticles, "SimParticles"); err != nil {
		t.Error("Can't add SimParticles collection: ", err)
	}
	part2 := &MCParticle{PDG: 11}
	SimParticles.Entries = append(SimParticles.Entries, part2)
	part3 := &MCParticle{PDG: 22}
	SimParticles.Entries = append(SimParticles.Entries, part3)

	part1.Children = append(part1.Children, eventOut.Reference(part2))
	part1.Children = append(part1.Children, eventOut.Reference(part3))
	part2.Parents = append(part2.Parents, eventOut.Reference(part1))
	part3.Parents = append(part3.Parents, eventOut.Reference(part1))

	writer.PushEvent(eventOut)

	reader := NewReader(buffer)

	eventIn, err := reader.Next()
	if err != nil {
		t.Error("Error reading back event")
	}

	MCParticles_, err := eventIn.Get("MCParticles")
	if err != nil {
		t.Error("Failed to get MCParticles collection")
	}

	part1_ := MCParticles_.GetEntry(0).(*MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match MCParticle")
	}
	part2_ := eventOut.Dereference(part1_.Children[0]).(*MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match first daughter particle")
	}
	part3_ := eventOut.Dereference(part1_.Children[1]).(*MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match second daughter particle")
	}
	part1_ = eventOut.Dereference(part2_.Parents[0]).(*MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of first daughter particle")
	}
	part1_ = eventOut.Dereference(part3_.Parents[0]).(*MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of second daughter particle")
	}

	if eventIn.Header.NUniqueIDs != eventOut.Header.NUniqueIDs {
		t.Error("Unique ID count was not carried over in push/get")
	}
}
