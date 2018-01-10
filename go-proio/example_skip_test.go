package proio_test

import (
	"bytes"
	"fmt"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model/lcio"
)

func Example_skip() {
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

	bytesReader := bytes.NewReader(buffer.Bytes())
	reader := proio.NewReader(bytesReader)

	reader.Skip(3)
	event, _ := reader.Next()
	fmt.Println(event)
	reader.SeekToStart()
	event, _ = reader.Next()
	fmt.Println(event)

	// Output:
    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle PDG:443 charge:4

    // Tag: Particles
    // ID:1 Type:proio.model.lcio.MCParticle PDG:443 charge:1
}
