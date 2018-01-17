package proio

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/decibelcooper/proio/go-proio/model/eic"
	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

func doWrite(writer *Writer, b *testing.B) {
	if b.N < 5000 {
		b.N = 5000
	}

	event := NewEvent()

	for i := 0; i < 1000; i++ {
		event.AddEntry("SimCaloHits", &prolcio.SimCalorimeterHit{
			Energy: rand.Float32(),
			Pos: []float32{
				rand.Float32(),
				rand.Float32(),
				rand.Float32(),
			},
		})
	}

	event.AddEntry("SimTrackHits", &prolcio.SimTrackerHit{
		EDep: rand.Float32(),
		Pos: []float64{
			rand.Float64(),
			rand.Float64(),
			rand.Float64(),
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Push(event)
	}
	writer.Flush()
}

func doRead(reader *Reader, b *testing.B) {
	b.ResetTimer()
	for event := range reader.ScanEvents() {
		trackHitID := event.TaggedEntries("SimTrackHits")[0]
		_ = event.GetEntry(trackHitID)
	}
}

func BenchmarkWriteUncomp(b *testing.B) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(UNCOMPRESSED)

	doWrite(writer, b)
}

func BenchmarkWriteLZ4(b *testing.B) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(LZ4)

	doWrite(writer, b)
}

func BenchmarkWriteGZIP(b *testing.B) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(GZIP)

	doWrite(writer, b)
}

func BenchmarkReadUncomp(b *testing.B) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(UNCOMPRESSED)
	doWrite(writer, b)

	reader := NewReader(buffer)
	doRead(reader, b)
}

func BenchmarkReadLZ4(b *testing.B) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(LZ4)
	doWrite(writer, b)

	reader := NewReader(buffer)
	doRead(reader, b)
}

func BenchmarkReadGZIP(b *testing.B) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(GZIP)
	doWrite(writer, b)

	reader := NewReader(buffer)
	doRead(reader, b)
}

func BenchmarkAddRemove100Entries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		event := NewEvent()
		for i := 0; i < 100; i++ {
			event.AddEntry("Particle", &eic.Particle{})
		}
		for i := 0; i < 100; i++ {
			event.RemoveEntry(uint64(i + 1))
		}
	}
}

func BenchmarkAddRemove1000Entries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		event := NewEvent()
		for i := 0; i < 1000; i++ {
			event.AddEntry("Particle", &eic.Particle{})
		}
		for i := 0; i < 1000; i++ {
			event.RemoveEntry(uint64(i + 1))
		}
	}
}

func BenchmarkAddRemove10000Entries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		event := NewEvent()
		for i := 0; i < 10000; i++ {
			event.AddEntry("Particle", &eic.Particle{})
		}
		for i := 0; i < 10000; i++ {
			event.RemoveEntry(uint64(i + 1))
		}
	}
}
