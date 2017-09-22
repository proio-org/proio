package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	"go-hep.org/x/hep/lcio"

	eicio "github.com/decibelcooper/eicio/go-eicio"
	"github.com/decibelcooper/eicio/go-eicio/model"
)

var (
	outFile = flag.String("o", "", "file to save output to")
	doGzip  = flag.Bool("g", false, "compress the stdout output with gzip")
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: lcio2eicio [options] <lcio-input-file>
options:
`,
	)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		log.Fatal("Invalid arguments")
	}

	lcioReader, err := lcio.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer lcioReader.Close()

	var eicioWriter *eicio.Writer
	if *outFile == "" {
		if *doGzip {
			eicioWriter = eicio.NewGzipWriter(os.Stdout)
		} else {
			eicioWriter = eicio.NewWriter(os.Stdout)
		}
	} else {
		eicioWriter, err = eicio.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer eicioWriter.Close()

	for lcioReader.Next() {
		lcioEvent := lcioReader.Event()
		eicioEvent := eicio.NewEvent()

		eicioEvent.Header.RunNumber = uint64(lcioEvent.RunNumber)
		eicioEvent.Header.EventNumber = uint64(lcioEvent.EventNumber)

		for i, collName := range lcioEvent.Names() {
			lcioColl := lcioEvent.Get(collName)

			var eicioColl eicio.Collection
			switch lcioColl.(type) {
			case *lcio.McParticleContainer:
				eicioColl = convertMCParticleCollection(lcioColl.(*lcio.McParticleContainer), &lcioEvent, uint32(i+1))
			case *lcio.SimTrackerHitContainer:
				eicioColl = convertSimTrackerHitCollection(lcioColl.(*lcio.SimTrackerHitContainer), &lcioEvent, uint32(i+1))
			case *lcio.TrackerRawDataContainer:
				eicioColl = convertTrackerRawDataCollection(lcioColl.(*lcio.TrackerRawDataContainer), &lcioEvent, uint32(i+1))
			case *lcio.TrackerDataContainer:
				eicioColl = convertTrackerDataCollection(lcioColl.(*lcio.TrackerDataContainer), &lcioEvent, uint32(i+1))
			case *lcio.TrackerHitContainer:
				eicioColl = convertTrackerHitCollection(lcioColl.(*lcio.TrackerHitContainer), &lcioEvent, uint32(i+1))
			case *lcio.TrackerPulseContainer:
				eicioColl = convertTrackerPulseCollection(lcioColl.(*lcio.TrackerPulseContainer), &lcioEvent, uint32(i+1))
			case *lcio.TrackerHitPlaneContainer:
				eicioColl = convertTrackerHitPlaneCollection(lcioColl.(*lcio.TrackerHitPlaneContainer), &lcioEvent, uint32(i+1))
			case *lcio.TrackerHitZCylinderContainer:
				eicioColl = convertTrackerHitZCylinderCollection(lcioColl.(*lcio.TrackerHitZCylinderContainer), &lcioEvent, uint32(i+1))
			case *lcio.TrackContainer:
				eicioColl = convertTrackCollection(lcioColl.(*lcio.TrackContainer), &lcioEvent, uint32(i+1))
			case *lcio.SimCalorimeterHitContainer:
				eicioColl = convertSimCalorimeterHitCollection(lcioColl.(*lcio.SimCalorimeterHitContainer), &lcioEvent, uint32(i+1))
			case *lcio.RawCalorimeterHitContainer:
				eicioColl = convertRawCalorimeterHitCollection(lcioColl.(*lcio.RawCalorimeterHitContainer), &lcioEvent, uint32(i+1))
			case *lcio.CalorimeterHitContainer:
				eicioColl = convertCalorimeterHitCollection(lcioColl.(*lcio.CalorimeterHitContainer), &lcioEvent, uint32(i+1))
			case *lcio.ClusterContainer:
				eicioColl = convertClusterCollection(lcioColl.(*lcio.ClusterContainer), &lcioEvent, uint32(i+1))
			case *lcio.RecParticleContainer:
				eicioColl = convertRecParticleCollection(lcioColl.(*lcio.RecParticleContainer), &lcioEvent, uint32(i+1))
			case *lcio.VertexContainer:
				eicioColl = convertVertexCollection(lcioColl.(*lcio.VertexContainer), &lcioEvent, uint32(i+1))
			case *lcio.RelationContainer:
				eicioColl = convertRelationCollection(lcioColl.(*lcio.RelationContainer), &lcioEvent, uint32(i+1))
			}

			if eicioColl != nil {
				if err := eicioEvent.Add(eicioColl, collName); err != nil {
					log.Fatal("Failed to add collection ", collName, ": ", err)
				}
			}
		}

		eicioWriter.Push(eicioEvent)
	}

	err = lcioReader.Err()
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func convertIntParams(intParams map[string][]int32) map[string]*model.IntParams {
	params := map[string]*model.IntParams{}
	for key, value := range intParams {
		params[key] = &model.IntParams{Array: value}
	}
	return params
}

func convertFloatParams(floatParams map[string][]float32) map[string]*model.FloatParams {
	params := map[string]*model.FloatParams{}
	for key, value := range floatParams {
		params[key] = &model.FloatParams{Array: value}
	}
	return params
}

func convertStringParams(stringParams map[string][]string) map[string]*model.StringParams {
	params := map[string]*model.StringParams{}
	for key, value := range stringParams {
		params[key] = &model.StringParams{Array: value}
	}
	return params
}

func convertParams(lcioParams lcio.Params) *model.Params {
	return &model.Params{
		Ints:    convertIntParams(lcioParams.Ints),
		Floats:  convertFloatParams(lcioParams.Floats),
		Strings: convertStringParams(lcioParams.Strings),
	}
}

func makeRef(entry interface{}, event *lcio.Event) *model.Reference {
	for i, collName := range event.Names() {
		collGen := event.Get(collName)

		j := 0
		found := false
		switch collGen.(type) {
		case *lcio.McParticleContainer:
			coll := collGen.(*lcio.McParticleContainer)
			for j, _ = range coll.Particles {
				if &coll.Particles[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackerRawDataContainer:
			coll := collGen.(*lcio.TrackerRawDataContainer)
			for j, _ = range coll.Data {
				if &coll.Data[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackerDataContainer:
			coll := collGen.(*lcio.TrackerDataContainer)
			for j, _ = range coll.Data {
				if &coll.Data[j] == entry {
					found = true
					break
				}
			}
		case *lcio.RawCalorimeterHitContainer:
			coll := collGen.(*lcio.RawCalorimeterHitContainer)
			for j, _ = range coll.Hits {
				if &coll.Hits[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackContainer:
			coll := collGen.(*lcio.TrackContainer)
			for j, _ = range coll.Tracks {
				if &coll.Tracks[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackerHitContainer:
			coll := collGen.(*lcio.TrackerHitContainer)
			for j, _ = range coll.Hits {
				if &coll.Hits[j] == entry {
					found = true
					break
				}
			}
		case *lcio.ClusterContainer:
			coll := collGen.(*lcio.ClusterContainer)
			for j, _ = range coll.Clusters {
				if &coll.Clusters[j] == entry {
					found = true
					break
				}
			}
		case *lcio.CalorimeterHitContainer:
			coll := collGen.(*lcio.CalorimeterHitContainer)
			for j, _ = range coll.Hits {
				if &coll.Hits[j] == entry {
					found = true
					break
				}
			}
		case *lcio.RecParticleContainer:
			coll := collGen.(*lcio.RecParticleContainer)
			for j, _ = range coll.Parts {
				if &coll.Parts[j] == entry {
					found = true
					break
				}
			}
		case *lcio.VertexContainer:
			coll := collGen.(*lcio.VertexContainer)
			for j, _ = range coll.Vtxs {
				if &coll.Vtxs[j] == entry {
					found = true
					break
				}
			}
		}

		if found {
			return &model.Reference{
				CollID:  uint32(i + 1),
				EntryID: uint32(j + 1),
			}
		}
	}
	return nil
}

func makeRefs(entries interface{}, event *lcio.Event) []*model.Reference {
	slice := reflect.ValueOf(entries)
	refs := make([]*model.Reference, 0)
	for i := 0; i < slice.Len(); i++ {
		ref := makeRef(slice.Index(i).Interface(), event)
		if ref != nil {
			refs = append(refs, ref)
		}
	}
	return refs
}

func convertMCParticleCollection(lcioColl *lcio.McParticleContainer, lcioEvent *lcio.Event, collID uint32) *model.MCParticleCollection {
	eicioColl := &model.MCParticleCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Particles {
		eicioEntry := &model.MCParticle{
			Id:        uint32(i + 1),
			Parents:   makeRefs(lcioEntry.Parents, lcioEvent),
			Children:  makeRefs(lcioEntry.Children, lcioEvent),
			PDG:       lcioEntry.PDG,
			GenStatus: lcioEntry.GenStatus,
			SimStatus: lcioEntry.SimStatus,
			Vertex:    lcioColl.Particles[i].Vertex[:],
			Time:      lcioEntry.Time,
			P:         lcioColl.Particles[i].P[:],
			Mass:      lcioEntry.Mass,
			Charge:    lcioEntry.Charge,
			PEndPoint: lcioColl.Particles[i].PEndPoint[:],
			Spin:      lcioColl.Particles[i].Spin[:],
			ColorFlow: lcioColl.Particles[i].ColorFlow[:],
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertSimTrackerHitCollection(lcioColl *lcio.SimTrackerHitContainer, lcioEvent *lcio.Event, collID uint32) *model.SimTrackerHitCollection {
	eicioColl := &model.SimTrackerHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &model.SimTrackerHit{
			Id:         uint32(i + 1),
			CellID0:    lcioEntry.CellID0,
			CellID1:    lcioEntry.CellID1,
			Pos:        lcioColl.Hits[i].Pos[:],
			EDep:       lcioEntry.EDep,
			Time:       lcioEntry.Time,
			Mc:         makeRef(lcioEntry.Mc, lcioEvent),
			P:          lcioColl.Hits[i].Momentum[:],
			PathLength: lcioEntry.PathLength,
			Quality:    lcioEntry.Quality,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func copyUint16SliceToUint32(origSlice []uint16) []uint32 {
	slice := make([]uint32, 0)
	for _, value := range origSlice {
		slice = append(slice, uint32(value))
	}
	return slice
}

func convertTrackerRawDataCollection(lcioColl *lcio.TrackerRawDataContainer, lcioEvent *lcio.Event, collID uint32) *model.TrackerRawDataCollection {
	eicioColl := &model.TrackerRawDataCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Data {
		eicioEntry := &model.TrackerRawData{
			Id:      uint32(i + 1),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			ADCs:    copyUint16SliceToUint32(lcioEntry.ADCs),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerDataCollection(lcioColl *lcio.TrackerDataContainer, lcioEvent *lcio.Event, collID uint32) *model.TrackerDataCollection {
	eicioColl := &model.TrackerDataCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Data {
		eicioEntry := &model.TrackerData{
			Id:      uint32(i + 1),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			Charges: lcioEntry.Charges,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerHitCollection(lcioColl *lcio.TrackerHitContainer, lcioEvent *lcio.Event, collID uint32) *model.TrackerHitCollection {
	eicioColl := &model.TrackerHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &model.TrackerHit{
			Id:      uint32(i + 1),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Pos:     lcioColl.Hits[i].Pos[:],
			Cov:     lcioColl.Hits[i].Cov[:],
			Type:    lcioEntry.Type,
			EDep:    lcioEntry.EDep,
			EDepErr: lcioEntry.EDepErr,
			Time:    lcioEntry.Time,
			Quality: lcioEntry.Quality,
			RawHits: makeRefs(lcioEntry.RawHits, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerPulseCollection(lcioColl *lcio.TrackerPulseContainer, lcioEvent *lcio.Event, collID uint32) *model.TrackerPulseCollection {
	eicioColl := &model.TrackerPulseCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Pulses {
		eicioEntry := &model.TrackerPulse{
			Id:      uint32(i + 1),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			Charge:  lcioEntry.Charge,
			Cov:     lcioColl.Pulses[i].Cov[:],
			Quality: lcioEntry.Quality,
			TPC:     makeRef(lcioEntry.TPC, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerHitPlaneCollection(lcioColl *lcio.TrackerHitPlaneContainer, lcioEvent *lcio.Event, collID uint32) *model.TrackerHitPlaneCollection {
	eicioColl := &model.TrackerHitPlaneCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &model.TrackerHitPlane{
			Id:      uint32(i + 1),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Type:    lcioEntry.Type,
			Pos:     lcioColl.Hits[i].Pos[:],
			U:       lcioColl.Hits[i].U[:],
			V:       lcioColl.Hits[i].V[:],
			DU:      lcioEntry.DU,
			DV:      lcioEntry.DV,
			EDep:    lcioEntry.EDep,
			EDepErr: lcioEntry.EDepErr,
			Time:    lcioEntry.Time,
			Quality: lcioEntry.Quality,
			RawHits: makeRefs(lcioEntry.RawHits, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerHitZCylinderCollection(lcioColl *lcio.TrackerHitZCylinderContainer, lcioEvent *lcio.Event, collID uint32) *model.TrackerHitZCylinderCollection {
	eicioColl := &model.TrackerHitZCylinderCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &model.TrackerHitZCylinder{
			Id:      uint32(i + 1),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Type:    lcioEntry.Type,
			Pos:     lcioColl.Hits[i].Pos[:],
			Center:  lcioColl.Hits[i].Center[:],
			DRPhi:   lcioEntry.DRPhi,
			DZ:      lcioEntry.DZ,
			EDep:    lcioEntry.EDep,
			EDepErr: lcioEntry.EDepErr,
			Time:    lcioEntry.Time,
			Quality: lcioEntry.Quality,
			RawHits: makeRefs(lcioEntry.RawHits, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackStates(lcioStates []lcio.TrackState) []*model.Track_TrackState {
	slice := make([]*model.Track_TrackState, 0)
	for _, state := range lcioStates {
		slice = append(slice, &model.Track_TrackState{
			Loc:   state.Loc,
			D0:    state.D0,
			Phi:   state.Phi,
			Omega: state.Omega,
			Z0:    state.Z0,
			TanL:  state.TanL,
			Cov:   state.Cov[:],
			Ref:   state.Ref[:],
		})
	}
	return slice
}

func convertTrackCollection(lcioColl *lcio.TrackContainer, lcioEvent *lcio.Event, collID uint32) *model.TrackCollection {
	eicioColl := &model.TrackCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Tracks {
		eicioEntry := &model.Track{
			Id:         uint32(i + 1),
			Type:       lcioEntry.Type,
			Chi2:       lcioEntry.Chi2,
			NDF:        lcioEntry.NdF,
			DEdx:       lcioEntry.DEdx,
			DEdxErr:    lcioEntry.DEdxErr,
			Radius:     lcioEntry.Radius,
			SubDetHits: lcioEntry.SubDetHits,
			States:     convertTrackStates(lcioEntry.States),
			Tracks:     makeRefs(lcioEntry.Tracks, lcioEvent),
			Hits:       makeRefs(lcioEntry.Hits, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertContribs(lcioContribs []lcio.Contrib, lcioEvent *lcio.Event) []*model.SimCalorimeterHit_Contrib {
	slice := make([]*model.SimCalorimeterHit_Contrib, 0)
	for _, contrib := range lcioContribs {
		slice = append(slice, &model.SimCalorimeterHit_Contrib{
			MCParticle: makeRef(contrib.Mc, lcioEvent),
			Energy:     contrib.Energy,
			Time:       contrib.Time,
			PDG:        contrib.PDG,
			StepPos:    contrib.StepPos[:],
		})
	}
	return slice
}

func convertSimCalorimeterHitCollection(lcioColl *lcio.SimCalorimeterHitContainer, lcioEvent *lcio.Event, collID uint32) *model.SimCalorimeterHitCollection {
	eicioColl := &model.SimCalorimeterHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &model.SimCalorimeterHit{
			Id:            uint32(i + 1),
			CellID0:       lcioEntry.CellID0,
			CellID1:       lcioEntry.CellID1,
			Energy:        lcioEntry.Energy,
			Pos:           lcioColl.Hits[i].Pos[:],
			Contributions: convertContribs(lcioEntry.Contributions, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertRawCalorimeterHitCollection(lcioColl *lcio.RawCalorimeterHitContainer, lcioEvent *lcio.Event, collID uint32) *model.RawCalorimeterHitCollection {
	eicioColl := &model.RawCalorimeterHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &model.RawCalorimeterHit{
			Id:        uint32(i + 1),
			CellID0:   lcioEntry.CellID0,
			CellID1:   lcioEntry.CellID1,
			Amplitude: lcioEntry.Amplitude,
			TimeStamp: lcioEntry.TimeStamp,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertCalorimeterHitCollection(lcioColl *lcio.CalorimeterHitContainer, lcioEvent *lcio.Event, collID uint32) *model.CalorimeterHitCollection {
	eicioColl := &model.CalorimeterHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		lcioRawHit := lcioEntry.Raw
		var rawHit *model.Reference
		if lcioRawHit != nil {
			rawHit = makeRef(lcioEntry.Raw.(*lcio.RawCalorimeterHit), lcioEvent)
		}

		eicioEntry := &model.CalorimeterHit{
			Id:        uint32(i + 1),
			CellID0:   lcioEntry.CellID0,
			CellID1:   lcioEntry.CellID1,
			Energy:    lcioEntry.Energy,
			EnergyErr: lcioEntry.EnergyErr,
			Time:      lcioEntry.Time,
			Pos:       lcioColl.Hits[i].Pos[:],
			Type:      lcioEntry.Type,
			Raw:       rawHit,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertParticleID(pid *lcio.ParticleID) *model.ParticleID {
	return &model.ParticleID{
		Likelihood: pid.Likelihood,
		Type:       pid.Type,
		PDG:        pid.PDG,
		AlgType:    pid.AlgType,
		Params:     pid.Params,
	}
}

func convertParticleIDs(lcioParticleIDs []lcio.ParticleID) []*model.ParticleID {
	slice := make([]*model.ParticleID, 0)
	for _, pid := range lcioParticleIDs {
		slice = append(slice, convertParticleID(&pid))
	}
	return slice
}

func convertClusterCollection(lcioColl *lcio.ClusterContainer, lcioEvent *lcio.Event, collID uint32) *model.ClusterCollection {
	eicioColl := &model.ClusterCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Clusters {
		eicioEntry := &model.Cluster{
			Id:         uint32(i + 1),
			Type:       lcioEntry.Type,
			Energy:     lcioEntry.Energy,
			EnergyErr:  lcioEntry.EnergyErr,
			Pos:        lcioColl.Clusters[i].Pos[:],
			PosErr:     lcioColl.Clusters[i].PosErr[:],
			Theta:      lcioEntry.Theta,
			Phi:        lcioEntry.Phi,
			DirErr:     lcioColl.Clusters[i].DirErr[:],
			Shape:      lcioColl.Clusters[i].Shape[:],
			PIDs:       convertParticleIDs(lcioEntry.PIDs),
			Clusters:   makeRefs(lcioEntry.Clusters, lcioEvent),
			Hits:       makeRefs(lcioEntry.Clusters, lcioEvent),
			Weights:    lcioColl.Clusters[i].Weights[:],
			SubDetEnes: lcioColl.Clusters[i].SubDetEnes[:],
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func findParticleID(pids []lcio.ParticleID, pid *lcio.ParticleID) int32 {
	for i, _ := range pids {
		if &pids[i] == pid {
			return int32(i)
		}
	}
	return -1
}

func convertRecParticleCollection(lcioColl *lcio.RecParticleContainer, lcioEvent *lcio.Event, collID uint32) *model.RecParticleCollection {
	eicioColl := &model.RecParticleCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Parts {
		eicioEntry := &model.RecParticle{
			Id:            uint32(i + 1),
			Type:          lcioEntry.Type,
			P:             lcioColl.Parts[i].P[:],
			Energy:        lcioEntry.Energy,
			Cov:           lcioColl.Parts[i].Cov[:],
			Mass:          lcioEntry.Mass,
			Charge:        lcioEntry.Charge,
			Ref:           lcioColl.Parts[i].Ref[:],
			PIDs:          convertParticleIDs(lcioEntry.PIDs),
			PIDUsed:       findParticleID(lcioEntry.PIDs, lcioEntry.PIDUsed),
			GoodnessOfPID: lcioEntry.GoodnessOfPID,
			Recs:          makeRefs(lcioEntry.Recs, lcioEvent),
			Tracks:        makeRefs(lcioEntry.Tracks, lcioEvent),
			Clusters:      makeRefs(lcioEntry.Clusters, lcioEvent),
			StartVtx:      makeRef(lcioEntry.StartVtx, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertVertexCollection(lcioColl *lcio.VertexContainer, lcioEvent *lcio.Event, collID uint32) *model.VertexCollection {
	eicioColl := &model.VertexCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Vtxs {
		eicioEntry := &model.Vertex{
			Id:      uint32(i + 1),
			Primary: lcioEntry.Primary,
			AlgType: lcioEntry.AlgType,
			Chi2:    lcioEntry.Chi2,
			Prob:    lcioEntry.Prob,
			Pos:     lcioColl.Vtxs[i].Pos[:],
			Cov:     lcioColl.Vtxs[i].Cov[:],
			Params:  lcioEntry.Params,
			RecPart: makeRef(lcioEntry.RecPart, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertRelationCollection(lcioColl *lcio.RelationContainer, lcioEvent *lcio.Event, collID uint32) *model.RelationCollection {
	eicioColl := &model.RelationCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Rels {
		eicioEntry := &model.Relation{
			Id:     uint32(i + 1),
			From:   makeRef(lcioEntry.From, lcioEvent),
			To:     makeRef(lcioEntry.To, lcioEvent),
			Weight: lcioEntry.Weight,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}
