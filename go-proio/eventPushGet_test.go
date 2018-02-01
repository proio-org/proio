package proio

import (
	"bytes"
	"testing"

	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

func TestUncompPushGet1(t *testing.T) {
	eventPushGet1(UNCOMPRESSED, t)
}

func TestUncompPushGet2(t *testing.T) {
	eventPushGet2(UNCOMPRESSED, t)
}

func TestUncompPushSkipGet1(t *testing.T) {
	eventPushSkipGet1(UNCOMPRESSED, t)
}

func TestUncompPushSkipGet2(t *testing.T) {
	eventPushSkipGet2(UNCOMPRESSED, t)
}

func TestLZ4PushGet1(t *testing.T) {
	eventPushGet1(LZ4, t)
}

func TestLZ4PushGet2(t *testing.T) {
	eventPushGet2(LZ4, t)
}

func TestLZ4PushSkipGet1(t *testing.T) {
	eventPushSkipGet1(LZ4, t)
}

func TestLZ4PushSkipGet2(t *testing.T) {
	eventPushSkipGet2(LZ4, t)
}

func TestGZIPPushGet1(t *testing.T) {
	eventPushGet1(GZIP, t)
}

func TestGZIPPushGet2(t *testing.T) {
	eventPushGet2(GZIP, t)
}

func TestGZIPPushSkipGet1(t *testing.T) {
	eventPushSkipGet1(GZIP, t)
}

func TestGZIPPushSkipGet2(t *testing.T) {
	eventPushSkipGet2(GZIP, t)
}

func eventPushGet1(comp Compression, t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(comp)

	var eventsOut [2]*Event

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

	event = NewEvent()
	event.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[1] = event
	writer.Close()

	reader := NewReader(buffer)
	defer reader.Close()

	for i := 0; i < 2; i++ {
		event, err := reader.Next()
		if err != nil {
			t.Error(err)
		}
		if event == nil {
			t.Errorf("Event %v failed to Get", i)
		}
		if event.String() != eventsOut[i].String() {
			t.Errorf("Event %v corrupted", i)
		}
	}
}

func eventPushGet2(comp Compression, t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(comp)

	var eventsOut [2]*Event

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
	writer.Flush()

	event = NewEvent()
	event.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[1] = event
	writer.Close()

	reader := NewReader(buffer)
	defer reader.Close()

	for i := 0; i < 2; i++ {
		event, err := reader.Next()
		if err != nil {
			t.Error(err)
		}
		if event == nil {
			t.Errorf("Event %v failed to Get", i)
		}
		if event.String() != eventsOut[i].String() {
			t.Errorf("Event %v corrupted", i)
		}
	}
}

func eventPushSkipGet1(comp Compression, t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(comp)

	var eventsOut [2]*Event

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

	event = NewEvent()
	event.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[1] = event
	writer.Close()

	reader := NewReader(buffer)
	defer reader.Close()
	reader.Skip(1)

	event, err := reader.Next()
	if err != nil {
		t.Error(err)
	}
	if event == nil {
		t.Errorf("Event %v failed to Get", 1)
	}
	if event.String() != eventsOut[1].String() {
		t.Errorf("Event %v corrupted", 1)
	}
}

func eventPushSkipGet2(comp Compression, t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(comp)

	var eventsOut [2]*Event

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
	writer.Flush()

	event = NewEvent()
	event.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event)
	eventsOut[1] = event
	writer.Close()

	reader := NewReader(buffer)
	defer reader.Close()
	reader.Skip(1)

	event, err := reader.Next()
	if err != nil {
		t.Error(err)
	}
	if event == nil {
		t.Errorf("Event %v failed to Get", 1)
	}
	if event.String() != eventsOut[1].String() {
		t.Errorf("Event %v corrupted", 1)
	}
}
