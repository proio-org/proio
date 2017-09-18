package eicio

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"go-hep.org/x/hep/lcio"

	"github.com/decibelcooper/eicio/go-eicio/model"
)

func TestEventPushGet(t *testing.T) {
	buffer := &bytes.Buffer{}

	writer := NewWriter(buffer)

	event0Out := NewEvent()

	MCParticles := &model.MCParticleCollection{}
	MCParticles.Entries = append(MCParticles.Entries, &model.MCParticle{})
	MCParticles.Entries = append(MCParticles.Entries, &model.MCParticle{})
	event0Out.Add(MCParticles, "MCParticles")

	simTrackHits := &model.SimTrackerHitCollection{}
	simTrackHits.Entries = append(simTrackHits.Entries, &model.SimTrackerHit{})
	simTrackHits.Entries = append(simTrackHits.Entries, &model.SimTrackerHit{})
	event0Out.Add(simTrackHits, "TrackerHits")

	writer.Push(event0Out)

	event1Out := NewEvent()

	simTrackHits = &model.SimTrackerHitCollection{}
	simTrackHits.Entries = append(simTrackHits.Entries, &model.SimTrackerHit{})
	simTrackHits.Entries = append(simTrackHits.Entries, &model.SimTrackerHit{})
	event1Out.Add(simTrackHits, "TrackerHits")

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

	MCParticles := &model.MCParticleCollection{}
	if err := eventOut.Add(MCParticles, "MCParticles"); err != nil {
		t.Error("Can't add MCParticles collection: ", err)
	}

	part1 := &model.MCParticle{PDG: 11}
	part2 := &model.MCParticle{PDG: 11}
	part3 := &model.MCParticle{PDG: 22}
	MCParticles.Entries = append(MCParticles.Entries, part1, part2, part3)

	part1.Children = append(part1.Children, eventOut.Reference(part2), eventOut.Reference(part3))
	part2.Parents = append(part2.Parents, eventOut.Reference(part1))
	part3.Parents = append(part3.Parents, eventOut.Reference(part1))

	writer.Push(eventOut)

	reader := NewReader(buffer)

	eventIn, err := reader.Get()
	if err != nil {
		t.Error("Error reading back event")
	}

	MCParticles_ := eventIn.Get("MCParticles")
	if MCParticles_ == nil {
		t.Error("Failed to get MCParticles collection")
	}

	part1_ := MCParticles_.GetEntry(0).(*model.MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match first model.MCParticle")
	}
	part2_ := eventOut.Dereference(part1_.Children[0]).(*model.MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match first daughter particle")
	}
	part3_ := eventOut.Dereference(part1_.Children[1]).(*model.MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match second daughter particle")
	}
	part1_ = eventOut.Dereference(part2_.Parents[0]).(*model.MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of first daughter particle")
	}
	part1_ = eventOut.Dereference(part3_.Parents[0]).(*model.MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of second daughter particle")
	}
}

func TestRefDeref2(t *testing.T) {
	event := NewEvent()

	MCParticles := &model.MCParticleCollection{}
	if err := event.Add(MCParticles, "MCParticles"); err != nil {
		t.Error("Can't add MCParticles collection: ", err)
	}

	part1 := &model.MCParticle{PDG: 11}
	part2 := &model.MCParticle{PDG: 11}
	part3 := &model.MCParticle{PDG: 22}
	MCParticles.Entries = append(MCParticles.Entries, part1, part2, part3)

	part1.Children = append(part1.Children, event.Reference(part2), event.Reference(part3))
	part2.Parents = append(part2.Parents, event.Reference(part1))
	part3.Parents = append(part3.Parents, event.Reference(part1))

	MCParticles_ := event.Get("MCParticles")
	if MCParticles_ == nil {
		t.Error("Failed to get MCParticles collection")
	}

	part1_ := MCParticles_.GetEntry(0).(*model.MCParticle)
	if part1_ != part1 {
		t.Error("Failed to match first model.MCParticle")
	}
	part2_ := event.Dereference(part1_.Children[0]).(*model.MCParticle)
	if part2_ != part2 {
		t.Error("Failed to match first daughter particle")
	}
	part3_ := event.Dereference(part1_.Children[1]).(*model.MCParticle)
	if part2_ != part2 {
		t.Error("Failed to match second daughter particle")
	}
	part1_ = event.Dereference(part2_.Parents[0]).(*model.MCParticle)
	if part1_ != part1 {
		t.Error("Failed to match parent of first daughter particle")
	}
	part1_ = event.Dereference(part3_.Parents[0]).(*model.MCParticle)
	if part1_ != part1 {
		t.Error("Failed to match parent of second daughter particle")
	}
}

func TestRefDeref3(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := NewWriter(buffer)

	eventOut := NewEvent()

	MCParticles := &model.MCParticleCollection{}
	if err := eventOut.Add(MCParticles, "MCParticles"); err != nil {
		t.Error("Can't add MCParticles collection: ", err)
	}
	part1 := &model.MCParticle{PDG: 11}
	MCParticles.Entries = append(MCParticles.Entries, part1)

	SimParticles := &model.MCParticleCollection{}
	if err := eventOut.Add(SimParticles, "SimParticles"); err != nil {
		t.Error("Can't add SimParticles collection: ", err)
	}
	part2 := &model.MCParticle{PDG: 11}
	part3 := &model.MCParticle{PDG: 22}
	SimParticles.Entries = append(SimParticles.Entries, part2, part3)

	part1.Children = append(part1.Children, eventOut.Reference(part2), eventOut.Reference(part3))
	part2.Parents = append(part2.Parents, eventOut.Reference(part1))
	part3.Parents = append(part3.Parents, eventOut.Reference(part1))

	writer.Push(eventOut)

	reader := NewReader(buffer)

	eventIn, err := reader.Get()
	if err != nil {
		t.Error("Error reading back event")
	}

	MCParticles_ := eventIn.Get("MCParticles")
	if MCParticles_ == nil {
		t.Error("Failed to get MCParticles collection")
	}

	part1_ := MCParticles_.GetEntry(0).(*model.MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match model.MCParticle")
	}
	part2_ := eventOut.Dereference(part1_.Children[0]).(*model.MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match first daughter particle")
	}
	part3_ := eventOut.Dereference(part1_.Children[1]).(*model.MCParticle)
	if part2_.String() != part2.String() {
		t.Error("Failed to match second daughter particle")
	}
	part1_ = eventOut.Dereference(part2_.Parents[0]).(*model.MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of first daughter particle")
	}
	part1_ = eventOut.Dereference(part3_.Parents[0]).(*model.MCParticle)
	if part1_.String() != part1.String() {
		t.Error("Failed to match parent of second daughter particle")
	}

	if eventIn.Header.NUniqueIDs != eventOut.Header.NUniqueIDs {
		t.Error("Unique ID count was not carried over in push/get")
	}
}

type TruthRelation struct {
	Truth *model.MCParticle
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
	filename := "../samples/largeSample.eicio"
	reader, err := Open(filename)
	if err != nil {
		b.Skip("Skipping tracking benchmark: missing input file ", filename)
	}

	b.ResetTimer()
	tracking(reader, b)
}

func BenchmarkTrackingGzip(b *testing.B) {
	filename := "../samples/largeSample.eicio.gz"
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
	for i := 0; i < b.N; i++ {
		event, err := reader.Get()
		if err != nil {
			b.N = i
			break
		}

		truthColl := event.Get("MCParticle").(*model.MCParticleCollection)
		trackColl := event.Get("Tracks").(*model.TrackCollection)

		// FIXME: boost back from crossing angle?

		var truthRelations []TruthRelation
		for i, truth := range truthColl.Entries {
			if truth.GenStatus != 1 || truth.Charge == float32(0) {
				continue
			}

			pNorm := normalizeVector(truth.P)
			eta := math.Atanh(pNorm[2])
			pT := math.Sqrt(truth.P[0]*truth.P[0] + truth.P[1]*truth.P[1])

			if pT > 0.5 {
				truthRelations = append(truthRelations, TruthRelation{
					Truth: truthColl.Entries[i],
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

		for _, track := range trackColl.Entries {
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
