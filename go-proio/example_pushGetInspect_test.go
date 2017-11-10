package proio_test

import (
	"bytes"
	"fmt"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model/lcio"
)

func Example_pushGetInspect() {
	buffer := &bytes.Buffer{}
	writer := proio.NewWriter(buffer)

	eventOut := proio.NewEvent()
	eventOut.SetEventNumber(1)

	// Build MCParticle collections
	// These must be added to the event before they can be automatically
	// referenced

	MCParticles, _ := eventOut.NewCollection("MCParticles", "lcio.MCParticle")
	parent := &lcio.MCParticle{PDG: 11}
	parentID, _ := MCParticles.AddEntry(parent)

	simParticles, _ := eventOut.NewCollection("SimParticles", "lcio.MCParticle")
	child1 := &lcio.MCParticle{PDG: 11}
	child2 := &lcio.MCParticle{PDG: 22}
	childIDs, _ := simParticles.AddEntries(child1, child2)

	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	writer.Push(eventOut)

	// Event created and serialized, now to deserialize and inspect

	reader := proio.NewReader(buffer)
	eventIn, _ := reader.Get()

	mcColl, _ := eventIn.Get("MCParticles")
	fmt.Print(mcColl.NEntries(), " MCParticle(s)...\n")
	for i, parentID := range mcColl.EntryIDs(true) {
		part := mcColl.GetEntry(parentID).(*lcio.MCParticle)
		fmt.Print(i, ". PDG: ", part.PDG, "\n")
		fmt.Print("  ", len(part.Children), " Children...\n")
		for j, childID := range part.Children {
			fmt.Print("  ", j, ". PDG: ", eventIn.GetEntry(childID).(*lcio.MCParticle).PDG, "\n")
		}
	}

	// Output:
	// 1 MCParticle(s)...
	// 0. PDG: 11
	//   2 Children...
	//   0. PDG: 11
	//   1. PDG: 22
}
