package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	eicio "github.com/decibelCooper/eicio/go-eicio"
	"go-hep.org/x/hep/lcio"
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: lcio2eicio [options] <lcio-input-file> <eicio-output-file>
	options:
	`,
	)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage

	flag.Parse()

	if flag.NArg() != 2 {
		printUsage()
		log.Fatal("Invalid arguments")
	}

	lcioReader, err := lcio.Open(flag.Arg(0))
	if err != nil {
		log.Fatal("Unable to open LCIO file:", err)
	}
	defer lcioReader.Close()

	eicioWriter, err := eicio.Create(flag.Arg(1))
	if err != nil {
		log.Fatal("Unable to create EICIO writer:", err)
	}
	defer eicioWriter.Close()

	for lcioReader.Next() {
		lcioEvent := lcioReader.Event()
		eicioEvent := eicio.NewEvent()

		for _, collName := range lcioEvent.Names() {
			lcioColl := lcioEvent.Get(collName)

			var eicioColl eicio.Message
			switch lcioColl.(type) {
			case *lcio.McParticleContainer:
				eicioColl = convertMCParticleCollection(lcioColl.(*lcio.McParticleContainer))
			}

			if eicioColl != nil {
				eicioEvent.AddCollection(eicioColl, collName)
			}
		}

		eicioWriter.PushEvent(eicioEvent)
	}
}

func convertIntParams(intParams map[string][]int32) map[string]*eicio.IntParams {
	params := map[string]*eicio.IntParams{}
	for key, value := range intParams {
		params[key].Array = value
	}
	return params
}

func convertFloatParams(floatParams map[string][]float32) map[string]*eicio.FloatParams {
	params := map[string]*eicio.FloatParams{}
	for key, value := range floatParams {
		params[key].Array = value
	}
	return params
}

func convertStringParams(stringParams map[string][]string) map[string]*eicio.StringParams {
	params := map[string]*eicio.StringParams{}
	for key, value := range stringParams {
		params[key].Array = value
	}
	return params
}

func convertParams(lcioParams lcio.Params) *eicio.Params {
	return &eicio.Params{
		Ints:    convertIntParams(lcioParams.Ints),
		Floats:  convertFloatParams(lcioParams.Floats),
		Strings: convertStringParams(lcioParams.Strings),
	}
}

func convertMCParticleCollection(lcioColl *lcio.McParticleContainer) *eicio.MCParticleCollection {
	eicioColl := &eicio.MCParticleCollection{
		Flags:  int32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}
	return eicioColl
}
