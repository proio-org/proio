package main

import (
	"bytes"
	"fmt"
	"github.com/decibelCooper/eicio/go-eicio"
	"log"
)

func main() {
	buffer := &bytes.Buffer{}

	writer := eicio.NewWriter(buffer)

	event0Out := eicio.NewEvent()
	event0Out.Header.Id = 1
	event0Out.Header.Description = "First test event"

	MC := &eicio.MCParticleCollection{}
	MC.Particle = append(MC.Particle, &eicio.MCParticle{})
	MC.Particle = append(MC.Particle, &eicio.MCParticle{})
	event0Out.AddCollection(MC, "MCParticles")

	simTrack := &eicio.SimTrackerHitCollection{}
	simTrack.Hit = append(simTrack.Hit, &eicio.SimTrackerHit{})
	simTrack.Hit = append(simTrack.Hit, &eicio.SimTrackerHit{})
	event0Out.AddCollection(simTrack, "TrackerHits")

	writer.PushEvent(event0Out)

	event1Out := eicio.NewEvent()
	event1Out.Header.Id = 2
	event1Out.Header.Description = "Second test event"

	simTrack = &eicio.SimTrackerHitCollection{}
	simTrack.Hit = append(simTrack.Hit, &eicio.SimTrackerHit{})
	simTrack.Hit = append(simTrack.Hit, &eicio.SimTrackerHit{})
	event1Out.AddCollection(simTrack, "TrackerHits")

	writer.PushEvent(event1Out)

	reader := eicio.NewReader(buffer)

	event0In, err := reader.GetEvent()
	if err != nil {
		log.Fatal("Failed to read event")
	}
	fmt.Print(event0In)

	event1In, err := reader.GetEvent()
	if err != nil {
		log.Fatal("Failed to read event")
	}
	fmt.Print(event1In)
}
