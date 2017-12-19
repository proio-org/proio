package proio

import (
	"bytes"
	"io"
	"sync"
	"testing"

	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

func TestScan1(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	eventOut.AddEntries(
		"MCParticles",
		&prolcio.MCParticle{},
		&prolcio.MCParticle{},
	)
	eventOut.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	for i := 0; i < evtScanBufferSize*2; i++ {
		writer.Push(eventOut)
	}

	writer.Flush()

	reader := NewReader(buffer)
	evtsRead := make(chan int)
	done := make(chan int)
	var eventOutMutex sync.Mutex

	scanner := func() {
		for event := range reader.ScanEvents() {
			eventOutMutex.Lock()
			if event.String() != eventOut.String() {
				t.Error("Event corrupted")
			}
			eventOutMutex.Unlock()
			evtsRead <- 1
		}
		done <- 1
	}

	go scanner()
	go scanner()

	totEvtsRead := 0
	totDone := 0

waitLoop:
	for {
		select {
		case nEvts := <-evtsRead:
			totEvtsRead += nEvts
		case nDone := <-done:
			totDone += nDone
			if totDone == 2 {
				break waitLoop
			}
		}
	}

	if totEvtsRead != evtScanBufferSize*2 {
		t.Errorf("%v events read instead of %v", totEvtsRead, evtScanBufferSize*2)
	}

errLoop:
	for {
		select {
		case err := <-reader.Err:
			if err != io.EOF {
				t.Error(err)
			}
		default:
			break errLoop
		}
	}
}

func TestScan2(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	eventOut.AddEntries(
		"MCParticles",
		&prolcio.MCParticle{},
		&prolcio.MCParticle{},
	)
	eventOut.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	for i := 0; i < evtScanBufferSize*2; i++ {
		writer.Push(eventOut)
		if i%10 == 9 {
			writer.Flush()
		}
	}

	writer.Flush()

	reader := NewReader(buffer)
	evtsRead := make(chan int)
	done := make(chan int)
	var eventOutMutex sync.Mutex

	scanner := func() {
		for event := range reader.ScanEvents() {
			eventOutMutex.Lock()
			if event.String() != eventOut.String() {
				t.Error("Event corrupted")
			}
			eventOutMutex.Unlock()
			evtsRead <- 1
		}
		done <- 1
	}

	go scanner()
	go scanner()

	totEvtsRead := 0
	totDone := 0

waitLoop:
	for {
		select {
		case nEvts := <-evtsRead:
			totEvtsRead += nEvts
		case nDone := <-done:
			totDone += nDone
			if totDone == 2 {
				break waitLoop
			}
		}
	}

	if totEvtsRead != evtScanBufferSize*2 {
		t.Errorf("%v events read instead of %v", totEvtsRead, evtScanBufferSize*2)
	}

errLoop:
	for {
		select {
		case err := <-reader.Err:
			if err != io.EOF {
				t.Error(err)
			}
		default:
			break errLoop
		}
	}
}

func TestScan3(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	eventOut.AddEntries(
		"MCParticles",
		&prolcio.MCParticle{},
		&prolcio.MCParticle{},
	)
	eventOut.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	for i := 0; i < evtScanBufferSize*2; i++ {
		writer.Push(eventOut)
	}

	writer.Flush()

	reader := NewReader(buffer)
	evtsRead := make(chan int)
	done := make(chan int)
	var eventOutMutex sync.Mutex

	scanner := func() {
		for event := range reader.ScanEvents() {
			eventOutMutex.Lock()
			if event.String() != eventOut.String() {
				t.Error("Event corrupted")
			}
			eventOutMutex.Unlock()
			evtsRead <- 1
			if nSkipped, _ := reader.Skip(1); nSkipped == 1 {
				evtsRead <- 1
			}
		}
		done <- 1
	}

	go scanner()
	go scanner()

	totEvtsRead := 0
	totDone := 0

waitLoop:
	for {
		select {
		case nEvts := <-evtsRead:
			totEvtsRead += nEvts
		case nDone := <-done:
			totDone += nDone
			if totDone == 2 {
				break waitLoop
			}
		}
	}

	if totEvtsRead != evtScanBufferSize*2 {
		t.Errorf("%v events read instead of %v", totEvtsRead, evtScanBufferSize*2)
	}

errLoop:
	for {
		select {
		case err := <-reader.Err:
			if err != io.EOF {
				t.Error(err)
			}
		default:
			break errLoop
		}
	}
}

func TestScan4(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	eventOut.AddEntries(
		"MCParticles",
		&prolcio.MCParticle{},
		&prolcio.MCParticle{},
	)
	eventOut.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	for i := 0; i < evtScanBufferSize*2; i++ {
		writer.Push(eventOut)
		if i%10 == 9 {
			writer.Flush()
		}
	}

	writer.Flush()

	reader := NewReader(buffer)
	evtsRead := make(chan int)
	done := make(chan int)
	var eventOutMutex sync.Mutex

	scanner := func() {
		for event := range reader.ScanEvents() {
			eventOutMutex.Lock()
			if event.String() != eventOut.String() {
				t.Error("Event corrupted")
			}
			eventOutMutex.Unlock()
			evtsRead <- 1
			if nSkipped, _ := reader.Skip(1); nSkipped == 1 {
				evtsRead <- 1
			}
		}
		done <- 1
	}

	go scanner()
	go scanner()

	totEvtsRead := 0
	totDone := 0

waitLoop:
	for {
		select {
		case nEvts := <-evtsRead:
			totEvtsRead += nEvts
		case nDone := <-done:
			totDone += nDone
			if totDone == 2 {
				break waitLoop
			}
		}
	}

	if totEvtsRead != evtScanBufferSize*2 {
		t.Errorf("%v events read instead of %v", totEvtsRead, evtScanBufferSize*2)
	}

errLoop:
	for {
		select {
		case err := <-reader.Err:
			if err != io.EOF {
				t.Error(err)
			}
		default:
			break errLoop
		}
	}
}

func TestScan5(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	eventOut.AddEntries(
		"MCParticles",
		&prolcio.MCParticle{},
		&prolcio.MCParticle{},
	)
	eventOut.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	for i := 0; i < evtScanBufferSize*3; i++ {
		writer.Push(eventOut)
	}

	writer.Flush()

	reader := NewReader(buffer)
	evtsRead := make(chan int)
	done := make(chan int)
	var eventOutMutex sync.Mutex

	scanner := func() {
		for event := range reader.ScanEvents() {
			eventOutMutex.Lock()
			if event.String() != eventOut.String() {
				t.Error("Event corrupted")
			}
			eventOutMutex.Unlock()
			evtsRead <- 1
		}
		done <- 1
	}

	go scanner()
	go scanner()

	totEvtsRead := 0
	totDone := 0

waitLoop:
	for {
		select {
		case nEvts := <-evtsRead:
			if totEvtsRead == 0 {
				reader.StopScan()
			}
			totEvtsRead += nEvts
		case nDone := <-done:
			totDone += nDone
			if totDone == 2 {
				break waitLoop
			}
		}
	}

	if totEvtsRead >= evtScanBufferSize*3 {
		t.Errorf("Failed to stop scans")
	}

errLoop:
	for {
		select {
		case err := <-reader.Err:
			if err != io.EOF {
				t.Error(err)
			}
		default:
			break errLoop
		}
	}
}
