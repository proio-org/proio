package proio_test

import (
	"bytes"
	"fmt"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model/eic"
)

func Example_pushGetInspect() {
	buffer := &bytes.Buffer{}
	writer := proio.NewWriter(buffer)

	eventOut := proio.NewEvent()

	// Create entries and hold onto their IDs for referencing

	parentPDG := int32(443)
	parent := &eic.Particle{Pdg: &parentPDG}
	parentID := eventOut.AddEntry("Particle", parent)
	eventOut.TagEntry(parentID, "MC", "Primary")

	child1PDG := int32(11)
	child1 := &eic.Particle{Pdg: &child1PDG}
	child2PDG := int32(-11)
	child2 := &eic.Particle{Pdg: &child2PDG}
	childIDs := eventOut.AddEntries("Particle", child1, child2)
	for _, id := range childIDs {
		eventOut.TagEntry(id, "MC", "GenStable")
	}

	parent.Child = append(parent.Child, childIDs...)
	child1.Parent = append(child1.Parent, parentID)
	child2.Parent = append(child2.Parent, parentID)

	writer.Push(eventOut)

	writer.Flush()

	// Event created and serialized, now to deserialize and inspect

	reader := proio.NewReader(buffer)
	eventIn, _ := reader.Next()

	mcParts := eventIn.TaggedEntries("Primary")
	fmt.Print(len(mcParts), " Primary particle(s)...\n")
	for i, parentID := range mcParts {
		part := eventIn.GetEntry(parentID).(*eic.Particle)
		fmt.Print(i, ". PDG: ", part.GetPdg(), "\n")
		fmt.Print("  ", len(part.Child), " children...\n")
		for j, childID := range part.Child {
			fmt.Print("  ", j, ". PDG: ", eventIn.GetEntry(childID).(*eic.Particle).GetPdg(), "\n")
		}
	}

	// Output:
	// 1 Primary particle(s)...
	// 0. PDG: 443
	//   2 children...
	//   0. PDG: 11
	//   1. PDG: -11
}
