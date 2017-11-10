package proio // import "github.com/decibelcooper/proio/go-proio"

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"testing"

	"go-hep.org/x/hep/lcio"

	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

func TestEventPushGet(t *testing.T) {
	buffer := &bytes.Buffer{}

	writer := NewWriter(buffer)

	event0Out := NewEvent()

	MCParticles, err := event0Out.NewCollection("MCParticles", "lcio.MCParticle")
	if err != nil {
		t.Error(err)
	}
	if _, err := MCParticles.AddEntries(&prolcio.MCParticle{}, &prolcio.MCParticle{}); err != nil {
		t.Error(err)
	}

	simTrackHits, err := event0Out.NewCollection("TrackerHits", "lcio.SimTrackerHit")
	if err != nil {
		t.Error(err)
	}
	if _, err := simTrackHits.AddEntries(&prolcio.SimTrackerHit{}, &prolcio.SimTrackerHit{}); err != nil {
		t.Error(err)
	}

	writer.Push(event0Out)

	event1Out := NewEvent()

	simTrackHits, err = event1Out.NewCollection("TrackerHits", "lcio.SimTrackerHit")
	if err != nil {
		t.Error(err)
	}
	if _, err := simTrackHits.AddEntries(&prolcio.SimTrackerHit{}, &prolcio.SimTrackerHit{}); err != nil {
		t.Error(err)
	}

	writer.Push(event1Out)

	reader := NewReader(buffer)

	event0In, err := reader.Get()
	if err != nil {
		t.Error(err)
	}
	if event0In == nil {
		t.Error("Event 0 failed to Get")
	}
	if !reflect.DeepEqual(event0Out, event0In) {
		t.Error("Event 0 corrupted")
	}

	event1In, err := reader.Get()
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

	MCParticles, err := eventOut.NewCollection("MCParticles", "lcio.MCParticle")
	if err != nil {
		t.Error(err)
	}

	parent := &prolcio.MCParticle{PDG: 11}
	parentID, err := MCParticles.AddEntry(parent)
	if err != nil {
		t.Error(err)
	}
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: 22}
	childIDs, err := MCParticles.AddEntries(child1, child2)
	if err != nil {
		t.Error(err)
	}

	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	writer.Push(eventOut)

	reader := NewReader(buffer)

	eventIn, err := reader.Get()
	if err != nil {
		t.Error("Error reading back event: ", err)
	}

	MCParticles_, err := eventIn.Get("MCParticles")
	if err != nil {
		t.Error("Failed to get MCParticles collection")
	}

	parent_ := MCParticles_.GetEntry(MCParticles_.EntryIDs(true)[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match first prolcio.MCParticle")
	}
	child1_ := eventOut.GetEntry(parent_.Children[0]).(*prolcio.MCParticle)
	if child1_.String() != child1.String() {
		t.Error("Failed to match first daughter particle")
	}
	child2_ := eventOut.GetEntry(parent_.Children[1]).(*prolcio.MCParticle)
	if child1_.String() != child1.String() {
		t.Error("Failed to match second daughter particle")
	}
	parent_ = eventOut.GetEntry(child1_.Parents[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match parent of first daughter particle")
	}
	parent_ = eventOut.GetEntry(child2_.Parents[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match parent of second daughter particle")
	}
}

func TestRefDeref2(t *testing.T) {
	event := NewEvent()

	MCParticles, err := event.NewCollection("MCParticles", "lcio.MCParticle")
	if err != nil {
		t.Error(err)
	}

	parent := &prolcio.MCParticle{PDG: 11}
	parentID, err := MCParticles.AddEntry(parent)
	if err != nil {
		t.Error(err)
	}
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: 22}
	childIDs, err := MCParticles.AddEntries(child1, child2)
	if err != nil {
		t.Error(err)
	}

	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	MCParticles_, err := event.Get("MCParticles")
	if err != nil {
		t.Error("Failed to get MCParticles collection")
	}

	parent_ := MCParticles_.GetEntry(MCParticles_.EntryIDs(true)[0]).(*prolcio.MCParticle)
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

func TestRefDeref3(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()

	MCParticles, err := eventOut.NewCollection("MCParticles", "lcio.MCParticle")
	if err != nil {
		t.Error(err)
	}
	parent := &prolcio.MCParticle{PDG: 11}
	parentID, err := MCParticles.AddEntry(parent)
	if err != nil {
		t.Error(err)
	}

	simParticles, err := eventOut.NewCollection("SimParticles", "lcio.MCParticle")
	if err != nil {
		t.Error(err)
	}
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: 22}
	childIDs, err := simParticles.AddEntries(child1, child2)
	if err != nil {
		t.Error(err)
	}

	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	writer.Push(eventOut)

	reader := NewReader(buffer)

	eventIn, err := reader.Get()
	if err != nil {
		t.Error("Error reading back event: ", err)
	}

	MCParticles_, err := eventIn.Get("MCParticles")
	if err != nil {
		t.Error("Failed to get MCParticles collection")
	}

	parent_ := MCParticles_.GetEntry(MCParticles_.EntryIDs(true)[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match first prolcio.MCParticle")
	}
	child1_ := eventOut.GetEntry(parent_.Children[0]).(*prolcio.MCParticle)
	if child1_.String() != child1.String() {
		t.Error("Failed to match first daughter particle")
	}
	child2_ := eventOut.GetEntry(parent_.Children[1]).(*prolcio.MCParticle)
	if child1_.String() != child1.String() {
		t.Error("Failed to match second daughter particle")
	}
	parent_ = eventOut.GetEntry(child1_.Parents[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match parent of first daughter particle")
	}
	parent_ = eventOut.GetEntry(child2_.Parents[0]).(*prolcio.MCParticle)
	if parent_.String() != parent.String() {
		t.Error("Failed to match parent of second daughter particle")
	}
}

type TruthRelation struct {
	Truth *prolcio.MCParticle
	PNorm []float64
	Eta   float64
	P_T   float64
}

func normalizeVector(vector []float64) []float64 {
	normFactor := math.Sqrt(dotProduct(vector, vector))
	for i, value := range vector {
		vector[i] = value / normFactor
	}
	return vector
}

func dotProduct(vector1 []float64, vector2 []float64) float64 {
	return vector1[0]*vector2[0] + vector1[1]*vector2[1] + vector1[2]*vector2[2]
}

func BenchmarkTracking(b *testing.B) {
	filename := "../samples/largeSample.proio"
	reader, err := Open(filename)
	if err != nil {
		b.Skip("Skipping tracking benchmark: missing input file ", filename)
	}

	b.ResetTimer()
	tracking(reader, b)
}

func BenchmarkTrackingLZ4(b *testing.B) {
	filename := "../samples/largeSample.proio.lz4"
	reader, err := Open(filename)
	if err != nil {
		b.Skip("Skipping tracking benchmark: missing input file ", filename)
	}

	b.ResetTimer()
	tracking(reader, b)
}

func BenchmarkTrackingGzip(b *testing.B) {
	filename := "../samples/largeSample.proio.gz"
	reader, err := Open(filename)
	if err != nil {
		b.Skip("Skipping tracking benchmark: missing input file ", filename)
	}

	b.ResetTimer()
	tracking(reader, b)
}

type TruthRelationLCIO struct {
	Truth *lcio.McParticle
	PNorm [3]float64
	Eta   float64
	P_T   float64
}

func normalizeVectorLCIO(vector [3]float64) [3]float64 {
	normFactor := math.Sqrt(dotProductLCIO(vector, vector))
	for i, value := range vector {
		vector[i] = value / normFactor
	}
	return vector
}

func dotProductLCIO(vector1 [3]float64, vector2 [3]float64) float64 {
	return vector1[0]*vector2[0] + vector1[1]*vector2[1] + vector1[2]*vector2[2]
}

func BenchmarkTrackingLCIO(b *testing.B) {
	filename := "../samples/largeSample.slcio"
	reader, err := lcio.Open(filename)
	if err != nil {
		b.Skip("Skipping tracking benchmark: missing input file ", filename)
	}

	b.ResetTimer()
	trackingLCIO(reader, b)
}

func tracking(reader *Reader, b *testing.B) {
	b.N = 0
	for event := range reader.ScanEvents() {
		b.N++

		truthColl, err := event.Get("MCParticle")
		if err != nil {
			b.Error(err)
		}
		trackColl, err := event.Get("Tracks")
		if err != nil {
			b.Error(err)
		}

		// FIXME: boost back from crossing angle?

		var truthRelations []TruthRelation
		for _, truthID := range truthColl.EntryIDs(false) {
			truth := truthColl.GetEntry(truthID).(*prolcio.MCParticle)
			if truth.GenStatus != 1 || truth.Charge == float32(0) {
				continue
			}

			pNorm := normalizeVector(truth.P)
			eta := math.Atanh(pNorm[2])
			pT := math.Sqrt(truth.P[0]*truth.P[0] + truth.P[1]*truth.P[1])

			if pT > 0.5 {
				truthRelations = append(truthRelations, TruthRelation{
					Truth: truth,
					PNorm: pNorm,
					Eta:   eta,
					P_T:   pT,
				})

				/*
					trueResults <- TrueResult{
						Eta: eta,
						P_T: pT,
					}
				*/
			}
		}

		for _, trackID := range trackColl.EntryIDs(false) {
			track := trackColl.GetEntry(trackID).(*prolcio.Track)
			tanLambda := float64(track.States[0].TanL)

			lambda := math.Atan(tanLambda)
			px := math.Cos(float64(track.States[0].Phi)) * math.Cos(lambda)
			py := math.Sin(float64(track.States[0].Phi)) * math.Cos(lambda)
			pz := math.Sin(lambda)

			pNorm := [3]float64{px, py, pz}

			minAngle := math.Inf(1)
			minIndex := -1
			for i, truthRelation := range truthRelations {
				angle := math.Acos(dotProduct(pNorm[:], truthRelation.PNorm))
				if angle < minAngle {
					minAngle = angle
					minIndex = i
				}
			}

			if minIndex >= 0 && minAngle < 0.01 {
				/*
					trackResults <- TrackResult{
						MinAngle: minAngle,
						Eta:      truthRelations[minIndex].Eta,
						P_T:      truthRelations[minIndex].P_T,
					}
				*/

				truthRelations = append(truthRelations[:minIndex], truthRelations[minIndex+1:]...)
			}
		}
	}

errLoop:
	for {
		select {
		case err := <-reader.Err:
			if err != io.EOF {
				b.Error(err)
			}
		default:
			break errLoop
		}
	}
}

func trackingLCIO(reader *lcio.Reader, b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !reader.Next() {
			b.N = i
			break
		}

		event := reader.Event()

		truthColl := event.Get("MCParticle").(*lcio.McParticleContainer)
		trackColl := event.Get("Tracks").(*lcio.TrackContainer)

		// FIXME: boost back from crossing angle?

		var truthRelations []TruthRelationLCIO
		for i, truth := range truthColl.Particles {
			if truth.GenStatus != 1 || truth.Charge == float32(0) {
				continue
			}

			pNorm := normalizeVectorLCIO(truth.P)
			eta := math.Atanh(pNorm[2])
			pT := math.Sqrt(truth.P[0]*truth.P[0] + truth.P[1]*truth.P[1])

			if pT > 0.5 {
				truthRelations = append(truthRelations, TruthRelationLCIO{
					Truth: &truthColl.Particles[i],
					PNorm: pNorm,
					Eta:   eta,
					P_T:   pT,
				})

				/*
					trueResults <- TrueResult{
						Eta: eta,
						P_T: pT,
					}
				*/
			}
		}

		for _, track := range trackColl.Tracks {
			tanLambda := track.TanL()
			//eta := -math.Log(math.Sqrt(1+tanLambda*tanLambda) - tanLambda)

			lambda := math.Atan(tanLambda)
			px := math.Cos(track.Phi()) * math.Cos(lambda)
			py := math.Sin(track.Phi()) * math.Cos(lambda)
			pz := math.Sin(lambda)

			pNorm := [3]float64{px, py, pz}

			minAngle := math.Inf(1)
			minIndex := -1
			for i, truthRelation := range truthRelations {
				angle := math.Acos(dotProductLCIO(pNorm, truthRelation.PNorm))
				if angle < minAngle {
					minAngle = angle
					minIndex = i
				}
			}

			if minIndex >= 0 && minAngle < 0.01 {
				/*
					trackResults <- TrackResult{
						MinAngle: minAngle,
						Eta:      truthRelations[minIndex].Eta,
						P_T:      truthRelations[minIndex].P_T,
					}
				*/

				truthRelations = append(truthRelations[:minIndex], truthRelations[minIndex+1:]...)
			}
		}
	}
}
