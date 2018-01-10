package proio_test

import (
	"fmt"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model/lcio"
)

func Example_print() {
	event := proio.NewEvent()

	parent := &lcio.MCParticle{PDG: 443}
	parentID := event.AddEntry("Particles", parent)
	event.TagEntry(parentID, "MC", "Primary")

	child1 := &lcio.MCParticle{PDG: 11}
	child2 := &lcio.MCParticle{PDG: -11}
	childIDs := event.AddEntries("Particles", child1, child2)
	for _, id := range childIDs {
		event.TagEntry(id, "MC", "Simulated")
	}

	parent.Children = append(parent.Children, childIDs...)
	child1.Parents = append(child1.Parents, parentID)
	child2.Parents = append(child2.Parents, parentID)

	fmt.Print(event)

	// Output:
    // Tag: MC
    // ID:1 Type:proio.model.lcio.MCParticle children:2 children:3 PDG:443
    // ID:2 Type:proio.model.lcio.MCParticle parents:1 PDG:11
    // ID:3 Type:proio.model.lcio.MCParticle parents:1 PDG:-11
    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle children:2 children:3 PDG:443
    // ID:2 Type:proio.model.lcio.MCParticle parents:1 PDG:11
    // ID:3 Type:proio.model.lcio.MCParticle parents:1 PDG:-11
    // Tag: Primary
    // ID:1 Type:proio.model.lcio.MCParticle children:2 children:3 PDG:443
    // Tag: Simulated
    // ID:2 Type:proio.model.lcio.MCParticle parents:1 PDG:11
    // ID:3 Type:proio.model.lcio.MCParticle parents:1 PDG:-11
}
