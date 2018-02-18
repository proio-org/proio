package proio

import (
	"testing"

	"github.com/decibelcooper/proio/go-proio/model/eic"
	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

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
