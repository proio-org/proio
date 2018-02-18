package proio

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
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
			t.Errorf("Event %v failed to Get", i)
		}
		if event.String() != eventsOut[i].String() {
			t.Errorf("Event %v corrupted", i)
		}
	}
}

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

func TestWriteLZ4IterateFile(t *testing.T) {
	writeIterateFile(LZ4, t)
}

func TestWriteGZIPIterateFile(t *testing.T) {
	writeIterateFile(GZIP, t)
}

func TestWriteUncompIterateFile(t *testing.T) {
	writeIterateFile(UNCOMPRESSED, t)
}

func writeIterateFile(comp Compression, t *testing.T) {
	nEvents := 5

	tmpDir, err := ioutil.TempDir("", "proiotest")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "writeIterateFile")

	writer, err := Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	writer.SetCompression(comp)
	event := NewEvent()
	for i := 0; i < nEvents; i++ {
		writer.Push(event)
	}
	writer.Close()

	nEvents = 0

	reader, err := Open(tmpFile)
	if err != nil {
		t.Error(err)
	}
	for range reader.ScanEvents() {
		nEvents++
	}
	reader.Close()

	if nEvents != 5 {
		t.Errorf("nEvents is %v instead of 5", nEvents)
	}
}

func TestCreateFileInEmptyDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "proiotest")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpDir = filepath.Join(tmpDir, "nonExistant")
	tmpFile := filepath.Join(tmpDir, "nonExistant")

	writer, err := Create(tmpFile)
	if err == nil {
		t.Errorf("No error thrown for creating file in non-existent directory.  Path is \"%v\"", tmpFile)
		writer.Close()
	}
}

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
