package proio_test

import (
	"bytes"
	"fmt"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model/eic"
)

func Example_skip() {
	buffer := &bytes.Buffer{}
	writer := proio.NewWriter(buffer)

	pdg := int32(443)
	for i := 0; i < 5; i++ {
		event := proio.NewEvent()
		charge := float32(i + 1)
		p := &eic.Particle{
			Pdg:    &pdg,
			Charge: &charge,
		}
		event.AddEntry("Particle", p)
		writer.Push(event)
	}
	writer.Flush()

	bytesReader := bytes.NewReader(buffer.Bytes())
	reader := proio.NewReader(bytesReader)

	reader.Skip(3)
	event, _ := reader.Next()
	fmt.Print(event)
	reader.SeekToStart()
	event, _ = reader.Next()
	fmt.Print(event)

	// Output:
	// ---------- TAG: Particle ----------
	// ID: 1
	// Entry type: proio.model.eic.Particle
	// pdg: 443
	// charge: 4
	//
	// ---------- TAG: Particle ----------
	// ID: 1
	// Entry type: proio.model.eic.Particle
	// pdg: 443
	// charge: 1
}
