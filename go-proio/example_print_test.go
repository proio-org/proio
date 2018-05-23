package proio_test

import (
	"fmt"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model/eic"
)

func Example_print() {
	event := proio.NewEvent()

	parentPDG := int32(443)
	parent := &eic.Particle{Pdg: &parentPDG}
	parentID := event.AddEntry("Particle", parent)
	event.TagEntry(parentID, "MC", "Primary")

	child1PDG := int32(11)
	child1 := &eic.Particle{Pdg: &child1PDG}
	child2PDG := int32(-11)
	child2 := &eic.Particle{Pdg: &child2PDG}
	childIDs := event.AddEntries("Particle", child1, child2)
	for _, id := range childIDs {
		event.TagEntry(id, "MC", "GenStable")
	}

	parent.Child = append(parent.Child, childIDs...)
	child1.Parent = append(child1.Parent, parentID)
	child2.Parent = append(child2.Parent, parentID)

	fmt.Print(event)

	// Output:
    // ---------- TAG: GenStable ----------
    // ID: 2
    // Entry type: proio.model.eic.Particle
    // parent: 1
    // pdg: 11                             
    //            
    // ID: 3      
    // Entry type: proio.model.eic.Particle
    // parent: 1
    // pdg: -11                            
    //      
    // ---------- TAG: MC ----------       
    // ID: 1     
    // Entry type: proio.model.eic.Particle
    // child: 2
    // child: 3
    // pdg: 443                            
    //           
    // ID: 2   
    // Entry type: proio.model.eic.Particle
    // parent: 1    
    // pdg: 11                                               
    //                            
    // ID: 3
    // Entry type: proio.model.eic.Particle
    // parent: 1
    // pdg: -11

    // ---------- TAG: Particle ----------
    // ID: 1
    // Entry type: proio.model.eic.Particle
    // child: 2
    // child: 3
    // pdg: 443

    // ID: 2
    // Entry type: proio.model.eic.Particle
    // parent: 1
    // pdg: 11

    // ID: 3
    // Entry type: proio.model.eic.Particle
    // parent: 1
    // pdg: -11

    // ---------- TAG: Primary ----------
    // ID: 1
    // Entry type: proio.model.eic.Particle
    // child: 2
    // child: 3
    // pdg: 443
}
