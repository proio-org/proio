import io
import pytest

import proio

def test_next_header1():
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        event = proio.Event()
        writer.push(event)
        writer.push(event)
        writer.flush()
        writer.push(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        header = reader.next_header()
        assert(header.nEvents == 2)
        assert(reader.__next__() != None)
        try:
            reader.__next__()
            assert(False)
        except StopIteration:
            pass

def test_next_header2():
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        event = proio.Event()
        writer.push(event)
        writer.push(event)
        writer.flush()
        writer.push(event)

    buf.seek(0, 0)

    with proio.Reader(fileobj = buf) as reader:
        assert(reader.__next__() != None)
        header = reader.next_header()
        assert(header.nEvents == 1)
        try:
            reader.__next__()
            assert(False)
        except StopIteration:
            pass
