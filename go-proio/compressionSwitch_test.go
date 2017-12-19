package proio

import (
	"bytes"
	"testing"

	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

func TestCompSwitch(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(UNCOMPRESSED)

	var eventsOut [7]*Event

	event := NewEvent()
	event.AddEntries(
		"MCParticles",
		&prolcio.MCParticle{},
		&prolcio.MCParticle{},
	)
	event.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[0] = event

	writer.SetCompression(LZ4)

	event = NewEvent()
	event.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[1] = event

	writer.SetCompression(GZIP)

	event = NewEvent()
	event.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[2] = event

	writer.SetCompression(UNCOMPRESSED)

	event = NewEvent()
	event.AddEntries(
		"CaloHits",
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[3] = event

	writer.SetCompression(GZIP)

	event = NewEvent()
	event.AddEntries(
		"Blah",
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[4] = event

	writer.SetCompression(LZ4)

	event = NewEvent()
	event.AddEntries(
		"Foo",
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[5] = event

	writer.SetCompression(UNCOMPRESSED)

	event = NewEvent()
	event.AddEntries(
		"Bar",
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[6] = event

	writer.Close()

	reader := NewReader(buffer)

	for i := 0; i < 7; i++ {
		event, err := reader.Next()
		if err != nil {
			t.Error(err)
		}
		if event == nil {
			t.Error("Event %v failed to Get", i)
		}
		if event.String() != eventsOut[i].String() {
			t.Error("Event %v corrupted", i)
		}
	}
}
