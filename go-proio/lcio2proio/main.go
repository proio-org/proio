package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/decibelcooper/proio/go-proio"
	prolcio "github.com/decibelcooper/proio/go-proio/model/lcio"
	"go-hep.org/x/hep/lcio"
)

var (
	outFile        = flag.String("o", "", "create file to save output to")
	compLevel      = flag.Int("c", 1, "compression level: 0 for uncompressed, 1 for LZ4 compression, 2 for GZIP compression")
	updateInterval = flag.Int("u", 5, "update interval in seconds (set to 0 to disable)")
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: lcio2proio [options] <lcio-input-file>
options:
`,
	)
	flag.PrintDefaults()
}

var refCache map[interface{}]uint64
var refsToFix []*uint64
var refSlicesToFix [][]uint64

var collNames map[uint32]string

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

	var proioWriter *proio.Writer
	if *outFile == "" {
		proioWriter = proio.NewWriter(os.Stdout)
	} else {
		proioWriter, err = proio.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	switch *compLevel {
	case 2:
		proioWriter.SetCompression(proio.GZIP)
	case 1:
		proioWriter.SetCompression(proio.LZ4)
	default:
		proioWriter.SetCompression(proio.UNCOMPRESSED)
	}
	defer proioWriter.Close()

	nEvents := 0
	checkpoint := time.Now()
	for lcioReader.Next() {
		lcioEvent := lcioReader.Event()
		proioEvent := proio.NewEvent()
		refCache = make(map[interface{}]uint64)
		collNames = make(map[uint32]string)

		for i, collName := range lcioEvent.Names() {
			lcioColl := lcioEvent.Get(collName)
			collNames[uint32(i+1)] = collName

			switch lcioColl.(type) {
			case *lcio.McParticleContainer:
				convertMCParticleCollection(lcioColl.(*lcio.McParticleContainer), &lcioEvent, proioEvent, collName)
			case *lcio.SimTrackerHitContainer:
				convertSimTrackerHitCollection(lcioColl.(*lcio.SimTrackerHitContainer), &lcioEvent, proioEvent, collName)
			case *lcio.TrackerRawDataContainer:
				convertTrackerRawDataCollection(lcioColl.(*lcio.TrackerRawDataContainer), &lcioEvent, proioEvent, collName)
			case *lcio.TrackerDataContainer:
				convertTrackerDataCollection(lcioColl.(*lcio.TrackerDataContainer), &lcioEvent, proioEvent, collName)
			case *lcio.TrackerHitContainer:
				convertTrackerHitCollection(lcioColl.(*lcio.TrackerHitContainer), &lcioEvent, proioEvent, collName)
			case *lcio.TrackerPulseContainer:
				convertTrackerPulseCollection(lcioColl.(*lcio.TrackerPulseContainer), &lcioEvent, proioEvent, collName)
			case *lcio.TrackerHitPlaneContainer:
				convertTrackerHitPlaneCollection(lcioColl.(*lcio.TrackerHitPlaneContainer), &lcioEvent, proioEvent, collName)
			case *lcio.TrackerHitZCylinderContainer:
				convertTrackerHitZCylinderCollection(lcioColl.(*lcio.TrackerHitZCylinderContainer), &lcioEvent, proioEvent, collName)
			case *lcio.TrackContainer:
				convertTrackCollection(lcioColl.(*lcio.TrackContainer), &lcioEvent, proioEvent, collName)
			case *lcio.SimCalorimeterHitContainer:
				convertSimCalorimeterHitCollection(lcioColl.(*lcio.SimCalorimeterHitContainer), &lcioEvent, proioEvent, collName)
			case *lcio.RawCalorimeterHitContainer:
				convertRawCalorimeterHitCollection(lcioColl.(*lcio.RawCalorimeterHitContainer), &lcioEvent, proioEvent, collName)
			case *lcio.CalorimeterHitContainer:
				convertCalorimeterHitCollection(lcioColl.(*lcio.CalorimeterHitContainer), &lcioEvent, proioEvent, collName)
			case *lcio.ClusterContainer:
				convertClusterCollection(lcioColl.(*lcio.ClusterContainer), &lcioEvent, proioEvent, collName)
			case *lcio.RecParticleContainer:
				convertRecParticleCollection(lcioColl.(*lcio.RecParticleContainer), &lcioEvent, proioEvent, collName)
			case *lcio.VertexContainer:
				convertVertexCollection(lcioColl.(*lcio.VertexContainer), &lcioEvent, proioEvent, collName)
			case *lcio.RelationContainer:
				convertRelationCollection(lcioColl.(*lcio.RelationContainer), &lcioEvent, proioEvent, collName)
			}
		}

		fixRefs(proioEvent)

		proioWriter.Push(proioEvent)
		nEvents++

		if *updateInterval > 0 {
			now := time.Now()
			if now.Sub(checkpoint) > time.Duration(*updateInterval)*time.Second {
				log.Println(nEvents, "events completed")
				checkpoint = now
			}
		}
	}

	err = lcioReader.Err()
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}

func fixRefs(event *proio.Event) {
	for _, slice := range refSlicesToFix {
		for i := range slice {
			slice[i] = fixRef(event, slice[i])
		}
	}

	for _, refPtr := range refsToFix {
		*refPtr = fixRef(event, *refPtr)
	}
	refsToFix = make([]*uint64, 0)
	refSlicesToFix = make([][]uint64, 0)
}

func fixRef(event *proio.Event, value uint64) uint64 {
	if value == 0 {
		return 0
	}
	collName := collNames[uint32(value&0xffffffff)]
	collEntry := (value >> 32) - 1
	return event.TaggedEntries(collName)[collEntry]
}

func makeRef(entry interface{}, event *lcio.Event) uint64 {
	if id, ok := refCache[entry]; ok {
		return id
	}

	for i, collName := range event.Names() {
		collGen := event.Get(collName)

		j := 0
		found := false
		switch collGen.(type) {
		case *lcio.McParticleContainer:
			coll := collGen.(*lcio.McParticleContainer)
			for j = range coll.Particles {
				if &coll.Particles[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackerRawDataContainer:
			coll := collGen.(*lcio.TrackerRawDataContainer)
			for j = range coll.Data {
				if &coll.Data[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackerDataContainer:
			coll := collGen.(*lcio.TrackerDataContainer)
			for j = range coll.Data {
				if &coll.Data[j] == entry {
					found = true
					break
				}
			}
		case *lcio.RawCalorimeterHitContainer:
			coll := collGen.(*lcio.RawCalorimeterHitContainer)
			for j = range coll.Hits {
				if &coll.Hits[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackContainer:
			coll := collGen.(*lcio.TrackContainer)
			for j = range coll.Tracks {
				if &coll.Tracks[j] == entry {
					found = true
					break
				}
			}
		case *lcio.TrackerHitContainer:
			coll := collGen.(*lcio.TrackerHitContainer)
			for j = range coll.Hits {
				if &coll.Hits[j] == entry {
					found = true
					break
				}
			}
		case *lcio.ClusterContainer:
			coll := collGen.(*lcio.ClusterContainer)
			for j = range coll.Clusters {
				if &coll.Clusters[j] == entry {
					found = true
					break
				}
			}
		case *lcio.CalorimeterHitContainer:
			coll := collGen.(*lcio.CalorimeterHitContainer)
			for j = range coll.Hits {
				if &coll.Hits[j] == entry {
					found = true
					break
				}
			}
		case *lcio.RecParticleContainer:
			coll := collGen.(*lcio.RecParticleContainer)
			for j = range coll.Parts {
				if &coll.Parts[j] == entry {
					found = true
					break
				}
			}
		case *lcio.VertexContainer:
			coll := collGen.(*lcio.VertexContainer)
			for j = range coll.Vtxs {
				if &coll.Vtxs[j] == entry {
					found = true
					break
				}
			}
		}

		if found {
			return uint64((i + 1)) + uint64((j+1)<<32)
		}
	}
	return 0
}

func makeRefs(entries interface{}, event *lcio.Event) []uint64 {
	slice := reflect.ValueOf(entries)
	refs := make([]uint64, 0)
	for i := 0; i < slice.Len(); i++ {
		ref := makeRef(slice.Index(i).Interface(), event)
		if ref != 0 {
			refs = append(refs, ref)
		}
	}
	return refs
}

func nilZeroFloat64Slice(slice []float64) []float64 {
	for _, value := range slice {
		if value != 0. {
			return slice
		}
	}
	return nil
}

func nilZeroFloat32Slice(slice []float32) []float32 {
	for _, value := range slice {
		if value != 0. {
			return slice
		}
	}
	return nil
}

func nilZeroInt32Slice(slice []int32) []int32 {
	for _, value := range slice {
		if value != 0 {
			return slice
		}
	}
	return nil
}

func convertMCParticleCollection(lcioColl *lcio.McParticleContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Particles {
		proioEntry := &prolcio.MCParticle{
			Parents:   makeRefs(lcioEntry.Parents, lcioEvent),
			Children:  makeRefs(lcioEntry.Children, lcioEvent),
			PDG:       lcioEntry.PDG,
			GenStatus: lcioEntry.GenStatus,
			SimStatus: lcioEntry.SimStatus,
			Vertex:    nilZeroFloat64Slice(lcioColl.Particles[i].Vertex[:]),
			Time:      lcioEntry.Time,
			P:         nilZeroFloat64Slice(lcioColl.Particles[i].P[:]),
			Mass:      lcioEntry.Mass,
			Charge:    lcioEntry.Charge,
			PEndPoint: nilZeroFloat64Slice(lcioColl.Particles[i].PEndPoint[:]),
			Spin:      nilZeroFloat32Slice(lcioColl.Particles[i].Spin[:]),
			ColorFlow: nilZeroInt32Slice(lcioColl.Particles[i].ColorFlow[:]),
		}

		proioEvent.AddEntry(collName, proioEntry)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Parents)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Children)
	}
}

func convertSimTrackerHitCollection(lcioColl *lcio.SimTrackerHitContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Hits {
		proioEntry := &prolcio.SimTrackerHit{
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

		proioEvent.AddEntry(collName, proioEntry)
		refsToFix = append(refsToFix, &proioEntry.Mc)
	}
}

func copyUint16SliceToUint32(origSlice []uint16) []uint32 {
	slice := make([]uint32, 0)
	for _, value := range origSlice {
		slice = append(slice, uint32(value))
	}
	return slice
}

func convertTrackerRawDataCollection(lcioColl *lcio.TrackerRawDataContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for _, lcioEntry := range lcioColl.Data {
		proioEntry := &prolcio.TrackerRawData{
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			ADCs:    copyUint16SliceToUint32(lcioEntry.ADCs),
		}

		proioEvent.AddEntry(collName, proioEntry)
	}
}

func convertTrackerDataCollection(lcioColl *lcio.TrackerDataContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for _, lcioEntry := range lcioColl.Data {
		proioEntry := &prolcio.TrackerData{
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			Charges: lcioEntry.Charges,
		}

		proioEvent.AddEntry(collName, proioEntry)
	}
}

func convertTrackerHitCollection(lcioColl *lcio.TrackerHitContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Hits {
		proioEntry := &prolcio.TrackerHit{
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

		proioEvent.AddEntry(collName, proioEntry)
		refSlicesToFix = append(refSlicesToFix, proioEntry.RawHits)
	}
}

func convertTrackerPulseCollection(lcioColl *lcio.TrackerPulseContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Pulses {
		proioEntry := &prolcio.TrackerPulse{
			CellID0: lcioEntry.CellID0,
			CellID1: lcioEntry.CellID1,
			Time:    lcioEntry.Time,
			Charge:  lcioEntry.Charge,
			Cov:     lcioColl.Pulses[i].Cov[:],
			Quality: lcioEntry.Quality,
			TPC:     makeRef(lcioEntry.TPC, lcioEvent),
		}

		proioEvent.AddEntry(collName, proioEntry)
		refsToFix = append(refsToFix, &proioEntry.TPC)
	}
}

func convertTrackerHitPlaneCollection(lcioColl *lcio.TrackerHitPlaneContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Hits {
		proioEntry := &prolcio.TrackerHitPlane{
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

		proioEvent.AddEntry(collName, proioEntry)
		refSlicesToFix = append(refSlicesToFix, proioEntry.RawHits)
	}
}

func convertTrackerHitZCylinderCollection(lcioColl *lcio.TrackerHitZCylinderContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Hits {
		proioEntry := &prolcio.TrackerHitZCylinder{
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

		proioEvent.AddEntry(collName, proioEntry)
		refSlicesToFix = append(refSlicesToFix, proioEntry.RawHits)
	}
}

func convertTrackStates(lcioStates []lcio.TrackState) []*prolcio.Track_TrackState {
	slice := make([]*prolcio.Track_TrackState, 0)
	for _, state := range lcioStates {
		slice = append(slice, &prolcio.Track_TrackState{
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

func convertTrackCollection(lcioColl *lcio.TrackContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for _, lcioEntry := range lcioColl.Tracks {
		proioEntry := &prolcio.Track{
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

		proioEvent.AddEntry(collName, proioEntry)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Tracks)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Hits)
	}
}

func convertContribs(lcioContribs []lcio.Contrib, lcioEvent *lcio.Event) []*prolcio.SimCalorimeterHit_Contrib {
	slice := make([]*prolcio.SimCalorimeterHit_Contrib, 0)
	for _, contrib := range lcioContribs {
		proioContrib := &prolcio.SimCalorimeterHit_Contrib{
			MCParticle: makeRef(contrib.Mc, lcioEvent),
			Energy:     contrib.Energy,
			Time:       contrib.Time,
			PDG:        contrib.PDG,
			StepPos:    nilZeroFloat32Slice(contrib.StepPos[:]),
		}
		slice = append(slice, proioContrib)
		refsToFix = append(refsToFix, &proioContrib.MCParticle)
	}
	return slice
}

func convertSimCalorimeterHitCollection(lcioColl *lcio.SimCalorimeterHitContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Hits {
		proioEntry := &prolcio.SimCalorimeterHit{
			CellID0:       lcioEntry.CellID0,
			CellID1:       lcioEntry.CellID1,
			Energy:        lcioEntry.Energy,
			Pos:           nilZeroFloat32Slice(lcioColl.Hits[i].Pos[:]),
			Contributions: convertContribs(lcioEntry.Contributions, lcioEvent),
		}

		proioEvent.AddEntry(collName, proioEntry)
	}
}

func convertRawCalorimeterHitCollection(lcioColl *lcio.RawCalorimeterHitContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for _, lcioEntry := range lcioColl.Hits {
		proioEntry := &prolcio.RawCalorimeterHit{
			CellID0:   lcioEntry.CellID0,
			CellID1:   lcioEntry.CellID1,
			Amplitude: lcioEntry.Amplitude,
			TimeStamp: lcioEntry.TimeStamp,
		}

		proioEvent.AddEntry(collName, proioEntry)
	}
}

func convertCalorimeterHitCollection(lcioColl *lcio.CalorimeterHitContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Hits {
		lcioRawHit := lcioEntry.Raw
		var rawHit uint64
		if lcioRawHit != nil {
			rawHit = makeRef(lcioEntry.Raw.(*lcio.RawCalorimeterHit), lcioEvent)
		}

		proioEntry := &prolcio.CalorimeterHit{
			CellID0:   lcioEntry.CellID0,
			CellID1:   lcioEntry.CellID1,
			Energy:    lcioEntry.Energy,
			EnergyErr: lcioEntry.EnergyErr,
			Time:      lcioEntry.Time,
			Pos:       lcioColl.Hits[i].Pos[:],
			Type:      lcioEntry.Type,
			Raw:       rawHit,
		}

		proioEvent.AddEntry(collName, proioEntry)
		refsToFix = append(refsToFix, &proioEntry.Raw)
	}
}

func convertParticleID(pid *lcio.ParticleID) *prolcio.ParticleID {
	return &prolcio.ParticleID{
		Likelihood: pid.Likelihood,
		Type:       pid.Type,
		PDG:        pid.PDG,
		AlgType:    pid.AlgType,
		Params:     pid.Params,
	}
}

func convertParticleIDs(lcioParticleIDs []lcio.ParticleID) []*prolcio.ParticleID {
	slice := make([]*prolcio.ParticleID, 0)
	for _, pid := range lcioParticleIDs {
		slice = append(slice, convertParticleID(&pid))
	}
	return slice
}

func convertClusterCollection(lcioColl *lcio.ClusterContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Clusters {
		proioEntry := &prolcio.Cluster{
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

		proioEvent.AddEntry(collName, proioEntry)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Clusters)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Hits)
	}
}

func findParticleID(pids []lcio.ParticleID, pid *lcio.ParticleID) int32 {
	for i := range pids {
		if &pids[i] == pid {
			return int32(i)
		}
	}
	return -1
}

func convertRecParticleCollection(lcioColl *lcio.RecParticleContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Parts {
		proioEntry := &prolcio.RecParticle{
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

		proioEvent.AddEntry(collName, proioEntry)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Recs)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Tracks)
		refSlicesToFix = append(refSlicesToFix, proioEntry.Clusters)
		refsToFix = append(refsToFix, &proioEntry.StartVtx)
	}
}

func convertVertexCollection(lcioColl *lcio.VertexContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for i, lcioEntry := range lcioColl.Vtxs {
		proioEntry := &prolcio.Vertex{
			Primary: lcioEntry.Primary,
			AlgType: lcioEntry.AlgType,
			Chi2:    lcioEntry.Chi2,
			Prob:    lcioEntry.Prob,
			Pos:     lcioColl.Vtxs[i].Pos[:],
			Cov:     lcioColl.Vtxs[i].Cov[:],
			Params:  lcioEntry.Params,
			RecPart: makeRef(lcioEntry.RecPart, lcioEvent),
		}

		proioEvent.AddEntry(collName, proioEntry)
		refsToFix = append(refsToFix, &proioEntry.RecPart)
	}
}

func convertRelationCollection(lcioColl *lcio.RelationContainer, lcioEvent *lcio.Event, proioEvent *proio.Event, collName string) {
	for _, lcioEntry := range lcioColl.Rels {
		proioEntry := &prolcio.Relation{
			From:   makeRef(lcioEntry.From, lcioEvent),
			To:     makeRef(lcioEntry.To, lcioEvent),
			Weight: lcioEntry.Weight,
		}

		proioEvent.AddEntry(collName, proioEntry)
		//refsToFix = append(refsToFix, &proioEntry.From)
		//refsToFix = append(refsToFix, &proioEntry.To)
	}
}
