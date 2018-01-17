package proio

import (
	"testing"

	eic "github.com/decibelcooper/proio/go-proio/model/eic"
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
