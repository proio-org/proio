package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

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

		eicioEvent.Header.RunNumber = uint64(lcioEvent.RunNumber)
		eicioEvent.Header.EventNumber = uint64(lcioEvent.EventNumber)

		for i, collName := range lcioEvent.Names() {
			lcioColl := lcioEvent.Get(collName)

			var eicioColl eicio.Identifiable
			switch lcioColl.(type) {
			case *lcio.McParticleContainer:
				eicioColl = convertMCParticleCollection(lcioColl.(*lcio.McParticleContainer), &lcioEvent, uint32(i))
			case *lcio.SimTrackerHitContainer:
				eicioColl = convertSimTrackerHitCollection(lcioColl.(*lcio.SimTrackerHitContainer), &lcioEvent, uint32(i))
			case *lcio.TrackerRawDataContainer:
				eicioColl = convertTrackerRawDataCollection(lcioColl.(*lcio.TrackerRawDataContainer), &lcioEvent, uint32(i))
			case *lcio.TrackerDataContainer:
				eicioColl = convertTrackerDataCollection(lcioColl.(*lcio.TrackerDataContainer), &lcioEvent, uint32(i))
			case *lcio.TrackerHitContainer:
				eicioColl = convertTrackerHitCollection(lcioColl.(*lcio.TrackerHitContainer), &lcioEvent, uint32(i))
			case *lcio.TrackerPulseContainer:
				eicioColl = convertTrackerPulseCollection(lcioColl.(*lcio.TrackerPulseContainer), &lcioEvent, uint32(i))
			case *lcio.TrackerHitPlaneContainer:
				eicioColl = convertTrackerHitPlaneCollection(lcioColl.(*lcio.TrackerHitPlaneContainer), &lcioEvent, uint32(i))
			case *lcio.TrackerHitZCylinderContainer:
				eicioColl = convertTrackerHitZCylinderCollection(lcioColl.(*lcio.TrackerHitZCylinderContainer), &lcioEvent, uint32(i))
			case *lcio.TrackContainer:
				eicioColl = convertTrackCollection(lcioColl.(*lcio.TrackContainer), &lcioEvent, uint32(i))
			case *lcio.SimCalorimeterHitContainer:
				eicioColl = convertSimCalorimeterHitCollection(lcioColl.(*lcio.SimCalorimeterHitContainer), &lcioEvent, uint32(i))
			case *lcio.RawCalorimeterHitContainer:
				eicioColl = convertRawCalorimeterHitCollection(lcioColl.(*lcio.RawCalorimeterHitContainer), &lcioEvent, uint32(i))
			case *lcio.CalorimeterHitContainer:
				eicioColl = convertCalorimeterHitCollection(lcioColl.(*lcio.CalorimeterHitContainer), &lcioEvent, uint32(i))
			case *lcio.ClusterContainer:
				eicioColl = convertClusterCollection(lcioColl.(*lcio.ClusterContainer), &lcioEvent, uint32(i))
			case *lcio.RecParticleContainer:
				eicioColl = convertRecParticleCollection(lcioColl.(*lcio.RecParticleContainer), &lcioEvent, uint32(i))
			case *lcio.VertexContainer:
				eicioColl = convertVertexCollection(lcioColl.(*lcio.VertexContainer), &lcioEvent, uint32(i))
			case *lcio.RelationContainer:
				eicioColl = convertRelationCollection(lcioColl.(*lcio.RelationContainer), &lcioEvent, uint32(i))
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
		params[key] = &eicio.IntParams{Array: value}
	}
	return params
}

func convertFloatParams(floatParams map[string][]float32) map[string]*eicio.FloatParams {
	params := map[string]*eicio.FloatParams{}
	for key, value := range floatParams {
		params[key] = &eicio.FloatParams{Array: value}
	}
	return params
}

func convertStringParams(stringParams map[string][]string) map[string]*eicio.StringParams {
	params := map[string]*eicio.StringParams{}
	for key, value := range stringParams {
		params[key] = &eicio.StringParams{Array: value}
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

func makeRef(entry interface{}, event *lcio.Event) *eicio.Reference {
	//	if entry.CanAddr() {
	//		entry = entry.Addr()
	//	}

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
			return &eicio.Reference{
				CollID:  uint32(i),
				EntryID: uint32(j),
			}
		}
	}
	return nil
}

func makeRefs(entries interface{}, event *lcio.Event) []*eicio.Reference {
	slice := reflect.ValueOf(entries)
	refs := make([]*eicio.Reference, 0)
	for i := 0; i < slice.Len(); i++ {
		ref := makeRef(slice.Index(i).Interface(), event)
		if ref != nil {
			refs = append(refs, ref)
		}
	}
	return refs
}

func convertMCParticleCollection(lcioColl *lcio.McParticleContainer, lcioEvent *lcio.Event, collID uint32) *eicio.MCParticleCollection {
	eicioColl := &eicio.MCParticleCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Particles {
		eicioEntry := &eicio.MCParticle{
			Id:        uint32(i),
			Parents:   makeRefs(lcioEntry.Parents, lcioEvent),
			Children:  makeRefs(lcioEntry.Children, lcioEvent),
			PDG:       lcioEntry.PDG,
			GenStatus: lcioEntry.GenStatus,
			SimStatus: lcioEntry.SimStatus,
			Vertex:    lcioEntry.Vertex[:],
			Time:      lcioEntry.Time,
			P:         lcioEntry.P[:],
			Mass:      lcioEntry.Mass,
			Charge:    lcioEntry.Charge,
			PEndPoint: lcioEntry.PEndPoint[:],
			Spin:      lcioEntry.Spin[:],
			ColorFlow: lcioEntry.ColorFlow[:],
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertSimTrackerHitCollection(lcioColl *lcio.SimTrackerHitContainer, lcioEvent *lcio.Event, collID uint32) *eicio.SimTrackerHitCollection {
	eicioColl := &eicio.SimTrackerHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &eicio.SimTrackerHit{
			Id:         uint32(i),
			CellID0:    lcioEntry.CellID0,
			CellID1:    lcioEntry.CellID1,
			Pos:        lcioEntry.Pos[:],
			EDep:       lcioEntry.EDep,
			Time:       lcioEntry.Time,
			Mc:         makeRef(lcioEntry.Mc, lcioEvent),
			P:          lcioEntry.Momentum[:],
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

func convertTrackerRawDataCollection(lcioColl *lcio.TrackerRawDataContainer, lcioEvent *lcio.Event, collID uint32) *eicio.TrackerRawDataCollection {
	eicioColl := &eicio.TrackerRawDataCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Data {
		eicioEntry := &eicio.TrackerRawData{
			Id:      uint32(i),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			ADCs:    copyUint16SliceToUint32(lcioEntry.ADCs),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerDataCollection(lcioColl *lcio.TrackerDataContainer, lcioEvent *lcio.Event, collID uint32) *eicio.TrackerDataCollection {
	eicioColl := &eicio.TrackerDataCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Data {
		eicioEntry := &eicio.TrackerData{
			Id:      uint32(i),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			Charges: lcioEntry.Charges,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerHitCollection(lcioColl *lcio.TrackerHitContainer, lcioEvent *lcio.Event, collID uint32) *eicio.TrackerHitCollection {
	eicioColl := &eicio.TrackerHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &eicio.TrackerHit{
			Id:      uint32(i),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Type:    lcioEntry.Type,
			Pos:     lcioEntry.Pos[:],
			Cov:     lcioEntry.Cov[:],
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

func convertTrackerPulseCollection(lcioColl *lcio.TrackerPulseContainer, lcioEvent *lcio.Event, collID uint32) *eicio.TrackerPulseCollection {
	eicioColl := &eicio.TrackerPulseCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Pulses {
		eicioEntry := &eicio.TrackerPulse{
			Id:      uint32(i),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			Charge:  lcioEntry.Charge,
			Cov:     lcioEntry.Cov[:],
			Quality: lcioEntry.Quality,
			TPC:     makeRef(lcioEntry.TPC, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertTrackerHitPlaneCollection(lcioColl *lcio.TrackerHitPlaneContainer, lcioEvent *lcio.Event, collID uint32) *eicio.TrackerHitPlaneCollection {
	eicioColl := &eicio.TrackerHitPlaneCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &eicio.TrackerHitPlane{
			Id:      uint32(i),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Type:    lcioEntry.Type,
			Pos:     lcioEntry.Pos[:],
			U:       lcioEntry.U[:],
			V:       lcioEntry.V[:],
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

func convertTrackerHitZCylinderCollection(lcioColl *lcio.TrackerHitZCylinderContainer, lcioEvent *lcio.Event, collID uint32) *eicio.TrackerHitZCylinderCollection {
	eicioColl := &eicio.TrackerHitZCylinderCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &eicio.TrackerHitZCylinder{
			Id:      uint32(i),
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Type:    lcioEntry.Type,
			Pos:     lcioEntry.Pos[:],
			Center:  lcioEntry.Center[:],
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

func convertTrackStates(lcioStates []lcio.TrackState) []*eicio.Track_TrackState {
	slice := make([]*eicio.Track_TrackState, 0)
	for _, state := range lcioStates {
		slice = append(slice, &eicio.Track_TrackState{
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

func convertTrackCollection(lcioColl *lcio.TrackContainer, lcioEvent *lcio.Event, collID uint32) *eicio.TrackCollection {
	eicioColl := &eicio.TrackCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Tracks {
		eicioEntry := &eicio.Track{
			Id:         uint32(i),
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

func convertContribs(lcioContribs []lcio.Contrib, lcioEvent *lcio.Event) []*eicio.SimCalorimeterHit_Contrib {
	slice := make([]*eicio.SimCalorimeterHit_Contrib, 0)
	for _, contrib := range lcioContribs {
		slice = append(slice, &eicio.SimCalorimeterHit_Contrib{
			MCParticle: makeRef(contrib.Mc, lcioEvent),
			Energy:     contrib.Energy,
			Time:       contrib.Time,
			PDG:        contrib.PDG,
			StepPos:    contrib.StepPos[:],
		})
	}
	return slice
}

func convertSimCalorimeterHitCollection(lcioColl *lcio.SimCalorimeterHitContainer, lcioEvent *lcio.Event, collID uint32) *eicio.SimCalorimeterHitCollection {
	eicioColl := &eicio.SimCalorimeterHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &eicio.SimCalorimeterHit{
			Id:            uint32(i),
			CellID0:       lcioEntry.CellID0,
			CellID1:       lcioEntry.CellID1,
			Energy:        lcioEntry.Energy,
			Pos:           lcioEntry.Pos[:],
			Contributions: convertContribs(lcioEntry.Contributions, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertRawCalorimeterHitCollection(lcioColl *lcio.RawCalorimeterHitContainer, lcioEvent *lcio.Event, collID uint32) *eicio.RawCalorimeterHitCollection {
	eicioColl := &eicio.RawCalorimeterHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		eicioEntry := &eicio.RawCalorimeterHit{
			Id:        uint32(i),
			CellID0:   lcioEntry.CellID0,
			CellID1:   lcioEntry.CellID1,
			Amplitude: lcioEntry.Amplitude,
			TimeStamp: lcioEntry.TimeStamp,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertCalorimeterHitCollection(lcioColl *lcio.CalorimeterHitContainer, lcioEvent *lcio.Event, collID uint32) *eicio.CalorimeterHitCollection {
	eicioColl := &eicio.CalorimeterHitCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Hits {
		lcioRawHit := lcioEntry.Raw
		var rawHit *eicio.Reference
		if lcioRawHit != nil {
			rawHit = makeRef(lcioEntry.Raw.(*lcio.RawCalorimeterHit), lcioEvent)
		}

		eicioEntry := &eicio.CalorimeterHit{
			Id:        uint32(i),
			CellID0:   lcioEntry.CellID0,
			CellID1:   lcioEntry.CellID1,
			Energy:    lcioEntry.Energy,
			EnergyErr: lcioEntry.EnergyErr,
			Time:      lcioEntry.Time,
			Pos:       lcioEntry.Pos[:],
			Type:      lcioEntry.Type,
			Raw:       rawHit,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertParticleID(pid *lcio.ParticleID) *eicio.ParticleID {
	return &eicio.ParticleID{
		Likelihood: pid.Likelihood,
		Type:       pid.Type,
		PDG:        pid.PDG,
		AlgType:    pid.AlgType,
		Params:     pid.Params,
	}
}

func convertParticleIDs(lcioParticleIDs []lcio.ParticleID) []*eicio.ParticleID {
	slice := make([]*eicio.ParticleID, 0)
	for _, pid := range lcioParticleIDs {
		slice = append(slice, convertParticleID(&pid))
	}
	return slice
}

func convertClusterCollection(lcioColl *lcio.ClusterContainer, lcioEvent *lcio.Event, collID uint32) *eicio.ClusterCollection {
	eicioColl := &eicio.ClusterCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Clusters {
		eicioEntry := &eicio.Cluster{
			Id:         uint32(i),
			Type:       lcioEntry.Type,
			Energy:     lcioEntry.Energy,
			EnergyErr:  lcioEntry.EnergyErr,
			Pos:        lcioEntry.Pos[:],
			PosErr:     lcioEntry.PosErr[:],
			Theta:      lcioEntry.Theta,
			Phi:        lcioEntry.Phi,
			DirErr:     lcioEntry.DirErr[:],
			Shape:      lcioEntry.Shape[:],
			PIDs:       convertParticleIDs(lcioEntry.PIDs),
			Clusters:   makeRefs(lcioEntry.Clusters, lcioEvent),
			Hits:       makeRefs(lcioEntry.Hits, lcioEvent),
			Weights:    lcioEntry.Weights[:],
			SubDetEnes: lcioEntry.SubDetEnes[:],
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

func convertRecParticleCollection(lcioColl *lcio.RecParticleContainer, lcioEvent *lcio.Event, collID uint32) *eicio.RecParticleCollection {
	eicioColl := &eicio.RecParticleCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Parts {
		eicioEntry := &eicio.RecParticle{
			Id:            uint32(i),
			Type:          lcioEntry.Type,
			P:             lcioEntry.P[:],
			Energy:        lcioEntry.Energy,
			Cov:           lcioEntry.Cov[:],
			Mass:          lcioEntry.Mass,
			Charge:        lcioEntry.Charge,
			Ref:           lcioEntry.Ref[:],
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

func convertVertexCollection(lcioColl *lcio.VertexContainer, lcioEvent *lcio.Event, collID uint32) *eicio.VertexCollection {
	eicioColl := &eicio.VertexCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Vtxs {
		eicioEntry := &eicio.Vertex{
			Id:      uint32(i),
			Primary: lcioEntry.Primary,
			AlgType: lcioEntry.AlgType,
			Chi2:    lcioEntry.Chi2,
			Prob:    lcioEntry.Prob,
			Pos:     lcioEntry.Pos[:],
			Cov:     lcioEntry.Cov[:],
			Params:  lcioEntry.Params,
			RecPart: makeRef(lcioEntry.RecPart, lcioEvent),
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}

func convertRelationCollection(lcioColl *lcio.RelationContainer, lcioEvent *lcio.Event, collID uint32) *eicio.RelationCollection {
	eicioColl := &eicio.RelationCollection{
		Id:     collID,
		Flags:  uint32(lcioColl.Flags),
		Params: convertParams(lcioColl.Params),
	}

	for i, lcioEntry := range lcioColl.Rels {
		eicioEntry := &eicio.Relation{
			Id:     uint32(i),
			From:   makeRef(lcioEntry.From, lcioEvent),
			To:     makeRef(lcioEntry.To, lcioEvent),
			Weight: lcioEntry.Weight,
		}

		eicioColl.Entries = append(eicioColl.Entries, eicioEntry)
	}

	return eicioColl
}
