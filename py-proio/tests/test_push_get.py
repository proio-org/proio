import io
import pytest

import proio
import proio.model.lcio as prolcio

def test_push_get_lz4():
    push_get(proio.LZ4)

def test_push_get_gzip():
    push_get(proio.GZIP)

def test_push_get_uncompressed():
    push_get(proio.UNCOMPRESSED)

def push_get(comp):
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        writer.set_compression(comp)

        event0_out = proio.Event()
        event0_out.add_entries(
                [prolcio.MCParticle(),
                    prolcio.MCParticle()],
                ['MCParticle']
                )
        event0_out.add_entries(
                [prolcio.SimTrackerHit(),
                    prolcio.SimTrackerHit()],
                ['TrackerHits']
                )
        writer.push(event0_out)
        
        event1_out = proio.Event()
        event1_out.add_entries(
                [prolcio.SimTrackerHit(),
                    prolcio.SimTrackerHit()],
                ['TrackerHits']
                )
        writer.push(event1_out)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        event0_in = reader.next()
        print(event0_in)
        assert event0_in != None
        assert event0_in.__str__() == event0_out.__str__()

        event1_in = reader.next()
        print(event1_in)
        assert event1_in != None
        assert event1_in.__str__() == event1_out.__str__()
