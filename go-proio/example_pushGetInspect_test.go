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

	// Create entries and hold onto their IDs for referencing

	parent := &lcio.MCParticle{PDG: 443}
	parentID := eventOut.AddEntry("Particles", parent)
	eventOut.TagEntry(parentID, "MC", "Primary")

	child1 := &lcio.MCParticle{PDG: 11}
	child2 := &lcio.MCParticle{PDG: -11}
	childIDs := eventOut.AddEntries("Particles", child1, child2)
	for _, id := range childIDs {
		eventOut.TagEntry(id, "MC", "Simulated")
	}

	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	writer.Push(eventOut)

	writer.Flush()

	// Event created and serialized, now to deserialize and inspect

	reader := proio.NewReader(buffer)
	eventIn, _ := reader.Next()

	mcParts := eventIn.TaggedEntries("Primary")
	fmt.Print(len(mcParts), " Primary particle(s)...\n")
	for i, parentID := range mcParts {
		part := eventIn.GetEntry(parentID).(*lcio.MCParticle)
		fmt.Print(i, ". PDG: ", part.PDG, "\n")
		fmt.Print("  ", len(part.Children), " children...\n")
		for j, childID := range part.Children {
			fmt.Print("  ", j, ". PDG: ", eventIn.GetEntry(childID).(*lcio.MCParticle).PDG, "\n")
		}
	}

	// Output:
	// 1 Primary particle(s)...
	// 0. PDG: 443
	//   2 children...
	//   0. PDG: 11
	//   1. PDG: -11
}
