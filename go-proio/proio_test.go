package proio // import "github.com/decibelcooper/proio/go-proio"

import (
	"bytes"
	"math"
	"testing"

	"go-hep.org/x/hep/lcio"

	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
)

func TestUncompPushGet(t *testing.T) {
	eventPushGet(UNCOMPRESSED, t)
}

func TestLZ4PushGet(t *testing.T) {
	eventPushGet(LZ4, t)
}

func TestGZIPPushGet(t *testing.T) {
	eventPushGet(GZIP, t)
}

func eventPushGet(comp Compression, t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)
	writer.SetCompression(comp)

	event0Out := NewEvent()
	event0Out.AddEntries(
		"MCParticles",
		&prolcio.MCParticle{},
		&prolcio.MCParticle{},
	)
	event0Out.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event0Out)
	writer.Flush()

	event1Out := NewEvent()
	event1Out.AddEntries(
		"TrackerHits",
		&prolcio.SimTrackerHit{},
		&prolcio.SimTrackerHit{},
	)
	writer.Push(event1Out)
	writer.Flush()

	reader := NewReader(buffer)

	event0In, err := reader.Next()
	if err != nil {
		t.Error(err)
	}
	if event0In == nil {
		t.Error("Event 0 failed to Get")
	}
	if event0Out.String() != event0In.String() {
		t.Error("Event 0 corrupted")
	}

	event1In, err := reader.Next()
	if err != nil {
		t.Error(err)
	}
	if event1In == nil {
		t.Error("Event 1 failed to Get")
	}
	if event1Out.String() != event1In.String() {
		t.Error("Event 1 corrupted")
	}
}

func TestRefDeref1(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	parent := &prolcio.MCParticle{PDG: 11}
	parentID := eventOut.AddEntry(parent, "MCParticles")
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: 22}
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
	parent := &prolcio.MCParticle{PDG: 11}
	parentID := event.AddEntry(parent, "MCParticles")
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: 22}
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

func TestRefDeref3(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()
	parent := &prolcio.MCParticle{PDG: 11}
	parentID := eventOut.AddEntry(parent, "MCParticles")
	child1 := &prolcio.MCParticle{PDG: 11}
	child2 := &prolcio.MCParticle{PDG: 22}
	childIDs := eventOut.AddEntries("SimParticles", child1, child2)
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
	filename := "repeatedSampleUncomp.proio"
	reader, err := Open(filename)
	if err != nil {
		b.Skip("Skipping tracking benchmark: missing input file ", filename)
	}

	b.ResetTimer()
	tracking(reader, b)
}

func BenchmarkTrackingLZ4(b *testing.B) {
	filename := "repeatedSample.proio"
	reader, err := Open(filename)
	if err != nil {
		b.Skip("Skipping tracking benchmark: missing input file ", filename)
	}

	b.ResetTimer()
	tracking(reader, b)
}

func BenchmarkTrackingGzip(b *testing.B) {
	filename := "repeatedSampleGZIP.proio"
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

		truthParts := event.TaggedEntries("MCParticle")
		tracks := event.TaggedEntries("Tracks")

		// FIXME: boost back from crossing angle?

		var truthRelations []TruthRelation
		for _, truthID := range truthParts {
			truth := event.GetEntry(truthID).(*prolcio.MCParticle)
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

		for _, trackID := range tracks {
			track := event.GetEntry(trackID).(*prolcio.Track)
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
