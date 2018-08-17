import io
import pytest

import proio
import proio.model.lcio as prolcio

def test_push_get1_lz4():
    push_get1(proio.LZ4)

def test_push_get2_lz4():
    push_get2(proio.LZ4)

def test_push_get3_lz4():
    push_get3(proio.LZ4)

def test_push_skip_get1_lz4():
    push_skip_get1(proio.LZ4)

def test_push_skip_get2_lz4():
    push_skip_get2(proio.LZ4)

def test_push_seek_skip_get1_lz4():
    push_seek_skip_get1(proio.LZ4)

def test_push_get1_gzip():
    push_get1(proio.GZIP)

def test_push_get2_gzip():
    push_get2(proio.GZIP)

def test_push_get3_gzip():
    push_get3(proio.GZIP)

def test_push_skip_get1_gzip():
    push_skip_get1(proio.GZIP)

def test_push_skip_get2_gzip():
    push_skip_get2(proio.GZIP)

def test_push_seek_skip_get1_gzip():
    push_seek_skip_get1(proio.GZIP)

def test_push_get1_uncompressed():
    push_get1(proio.UNCOMPRESSED)

def test_push_get2_uncompressed():
    push_get2(proio.UNCOMPRESSED)

def test_push_get3_uncompressed():
    push_get3(proio.UNCOMPRESSED)

def test_push_skip_get1_uncompressed():
    push_skip_get1(proio.UNCOMPRESSED)

def test_push_skip_get2_uncompressed():
    push_skip_get2(proio.UNCOMPRESSED)

def test_push_seek_skip_get1_uncompressed():
    push_seek_skip_get1(proio.UNCOMPRESSED)

def push_get1(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        eventsOut = []

        event = proio.Event()
        event.add_entries(
                'MCParticle',
                prolcio.MCParticle(),
                prolcio.MCParticle()
                )
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)
        
        event = proio.Event()
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        for i in range(0, len(eventsOut)):
            event = reader.__next__()
            assert event != None
            assert event.__str__() == eventsOut[i].__str__()

def push_get2(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        eventsOut = []

        event = proio.Event()
        event.add_entries(
                'MCParticle',
                prolcio.MCParticle(),
                prolcio.MCParticle()
                )
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)
        writer.flush()
        
        event = proio.Event()
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        for i in range(0, len(eventsOut)):
            event = reader.__next__()
            assert event != None
            assert event.__str__() == eventsOut[i].__str__()

def push_get3(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        event = proio.Event()
        event.add_entries(
                'MCParticle',
                prolcio.MCParticle(),
                prolcio.MCParticle()
                )
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        
        event = proio.Event()
        event.add_entry(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                )
        writer.push(event)

    buf.seek(0, 0)
    buf2 = io.BytesIO(b'')

    with proio.Reader(fileobj = buf) as reader:
        with proio.Writer(fileobj = buf2) as writer:
            eventsOut = []
            for event in reader:
                event.add_entry(
                        'TrackerHits',
                        prolcio.SimTrackerHit(),
                        )
                eventsOut.append(event)
                writer.push(event)

    buf2.seek(0, 0)

    with proio.Reader(fileobj = buf2) as reader:
        for i in range(0, len(eventsOut)):
            event = reader.__next__()
            assert event != None
            assert event.__str__() == eventsOut[i].__str__()

def push_skip_get1(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        eventsOut = []

        event = proio.Event()
        event.add_entries(
                'MCParticle',
                prolcio.MCParticle(),
                prolcio.MCParticle()
                )
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)
        
        event = proio.Event()
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        reader.skip(1)
        event = reader.__next__()
        assert event != None
        assert event.__str__() == eventsOut[1].__str__()

def push_skip_get2(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        eventsOut = []

        event = proio.Event()
        event.add_entries(
                'MCParticle',
                prolcio.MCParticle(),
                prolcio.MCParticle()
                )
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)
        writer.flush()
        
        event = proio.Event()
        event.add_entries(
                'TrackerHits',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        reader.skip(1)
        event = reader.__next__()
        assert event != None
        assert event.__str__() == eventsOut[1].__str__()

def push_seek_skip_get1(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        eventsOut = []

        event = proio.Event()
        event.add_entries(
                'Hit1',
                prolcio.MCParticle(),
                prolcio.MCParticle()
                )
        writer.push(event)
        eventsOut.append(event)
        
        event = proio.Event()
        event.add_entries(
                'Hit2',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)

        writer.flush()

        event = proio.Event()
        event.add_entries(
                'Hit3',
                prolcio.SimTrackerHit(),
                prolcio.SimTrackerHit()
                )
        writer.push(event)
        eventsOut.append(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        event = reader.__next__()
        assert event != None
        assert event.__str__() == eventsOut[0].__str__()

        reader.seek_to_start()
        reader.skip(2)
        event = reader.__next__()
        assert event != None
        assert event.__str__() == eventsOut[2].__str__()

