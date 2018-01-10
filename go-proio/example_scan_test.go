package proio_test

import (
	"bytes"
	"fmt"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model/lcio"
)

func Example_scan() {
	buffer := &bytes.Buffer{}
	writer := proio.NewWriter(buffer)

	for i := 0; i < 5; i++ {
		event := proio.NewEvent()
		p := &lcio.MCParticle{
			PDG:    443,
			Charge: float32(i + 1),
		}
		event.AddEntry("Particles", p)
		writer.Push(event)
	}
	writer.Flush()

	reader := proio.NewReader(buffer)

	for event := range reader.ScanEvents() {
		fmt.Println(event)
	}

	// Output:
    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle PDG:443 charge:1

    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle PDG:443 charge:2

    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle PDG:443 charge:3

    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle PDG:443 charge:4

    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle PDG:443 charge:5
}
