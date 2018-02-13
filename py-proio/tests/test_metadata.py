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
        assert(reader.next() != None)
        try:
            reader.next()
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
        assert(reader.next() != None)
        header = reader.next_header()
        assert(header.nEvents == 1)
        try:
            reader.next()
            assert(False)
        except StopIteration:
            pass
