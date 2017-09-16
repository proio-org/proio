package eicio_test

import (
	"bytes"
	"fmt"

	eicio "github.com/decibelcooper/eicio/go-eicio"
	"github.com/decibelcooper/eicio/go-eicio/model"
)

func Example_pushGetInspect() {
	buffer := &bytes.Buffer{}
	writer := eicio.NewWriter(buffer)

	eventOut := eicio.NewEvent()
	eventOut.Header.EventNumber = 1

	// Build MCParticle collections
	// These must be added to the event before they can be automatically
	// referenced

	MCParticles := &model.MCParticleCollection{}
	eventOut.Add(MCParticles, "MCParticles")
	part1 := &model.MCParticle{PDG: 11}
	MCParticles.Entries = append(MCParticles.Entries, part1)

	SimParticles := &model.MCParticleCollection{}
	eventOut.Add(SimParticles, "SimParticles")
	part2 := &model.MCParticle{PDG: 11}
	part3 := &model.MCParticle{PDG: 22}
	SimParticles.Entries = append(SimParticles.Entries, part2, part3)

	part1.Children = append(part1.Children, eventOut.Reference(part2), eventOut.Reference(part3))
	part2.Parents = append(part2.Parents, eventOut.Reference(part1))
	part3.Parents = append(part3.Parents, eventOut.Reference(part1))

	writer.Push(eventOut)

	// Event created and serialized, now to deserialize and inspect

	reader := eicio.NewReader(buffer)
	eventIn, _ := reader.Get()

	mcColl, _ := eventIn.Get("MCParticles").(*model.MCParticleCollection)
	fmt.Print(mcColl.GetNEntries(), " MCParticle(s)...\n")
	for i, part := range mcColl.Entries {
		fmt.Print(i, ". PDG: ", part.PDG, "\n")
		fmt.Print("  ", len(part.Children), " Children...\n")
		for j, ref := range part.Children {
			fmt.Print("  ", j, ". PDG: ", eventIn.Dereference(ref).(*model.MCParticle).PDG, "\n")
		}
	}

	// Output:
	// 1 MCParticle(s)...
	// 0. PDG: 11
	//   2 Children...
	//   0. PDG: 11
	//   1. PDG: 22
}
