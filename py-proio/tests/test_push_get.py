import io
import pytest

import proio
import proio.model.lcio as prolcio

def test_push_get1_lz4():
    push_get1(proio.LZ4)

def test_push_get2_lz4():
    push_get2(proio.LZ4)

def test_push_get1_gzip():
    push_get1(proio.GZIP)

def test_push_get2_gzip():
    push_get2(proio.GZIP)

def test_push_get1_uncompressed():
    push_get1(proio.UNCOMPRESSED)

def test_push_get2_uncompressed():
    push_get2(proio.UNCOMPRESSED)

def push_get1(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        eventsOut = []

        event = proio.Event()
        event.add_entries(
                [prolcio.MCParticle(),
                    prolcio.MCParticle()],
                ['MCParticle']
                )
        event.add_entries(
                [prolcio.SimTrackerHit(),
                    prolcio.SimTrackerHit()],
                ['TrackerHits']
                )
        writer.push(event)
        eventsOut.append(event)
        
        event = proio.Event()
        event.add_entries(
                [prolcio.SimTrackerHit(),
                    prolcio.SimTrackerHit()],
                ['TrackerHits']
                )
        writer.push(event)
        eventsOut.append(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        for i in range(0, len(eventsOut)):
            event = reader.next()
            assert event != None
            assert event.__str__() == eventsOut[i].__str__()

def push_get2(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        eventsOut = []

        event = proio.Event()
        event.add_entries(
                [prolcio.MCParticle(),
                    prolcio.MCParticle()],
                ['MCParticle']
                )
        event.add_entries(
                [prolcio.SimTrackerHit(),
                    prolcio.SimTrackerHit()],
                ['TrackerHits']
                )
        writer.push(event)
        eventsOut.append(event)
        writer.flush()
        
        event = proio.Event()
        event.add_entries(
                [prolcio.SimTrackerHit(),
                    prolcio.SimTrackerHit()],
                ['TrackerHits']
                )
        writer.push(event)
        eventsOut.append(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        for i in range(0, len(eventsOut)):
            event = reader.next()
            assert event != None
            assert event.__str__() == eventsOut[i].__str__()
