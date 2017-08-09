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
	event0Out.AddCollection(MCParticles, "MCParticles")

	simTrackHits := &SimTrackerHitCollection{}
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	event0Out.AddCollection(simTrackHits, "TrackerHits")

	writer.PushEvent(event0Out)

	event1Out := NewEvent()

	simTrackHits = &SimTrackerHitCollection{}
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	simTrackHits.Entries = append(simTrackHits.Entries, &SimTrackerHit{})
	event1Out.AddCollection(simTrackHits, "TrackerHits")

	writer.PushEvent(event1Out)

	reader := NewReader(buffer)

	event0In := reader.GetEvent()
	if event0In == nil {
		t.Error("Event 0 failed to Get")
	}
	if !reflect.DeepEqual(event0Out, event0In) {
		t.Error("Event 0 corrupted")
	}

	event1In := reader.GetEvent()
	if event1In == nil {
		t.Error("Event 1 failed to Get")
	}
	if !reflect.DeepEqual(event1Out, event1In) {
		t.Error("Event 1 corrupted")
	}
}
