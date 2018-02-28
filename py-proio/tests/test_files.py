import os
import tempfile
import threading

import proio

def test_write_iterate_file():
    n_events = 5
    test_file = tempfile.NamedTemporaryFile()

    with proio.Writer(fileobj = test_file) as writer:
        event = proio.Event()
        for i in range(0, n_events):
            writer.push(event)

    n_events = 0
    test_file.seek(0)

    with proio.Reader(fileobj = test_file) as reader:
        for event in reader:
            n_events += 1

    test_file.close()

    assert(n_events == 5)

def test_write_seek_file():
    test_file = tempfile.NamedTemporaryFile()

    with proio.Writer(fileobj = test_file) as writer:
        event = proio.Event()
        writer.push(event)
        writer.flush()
        writer.push(event)
        writer.push(event)
        writer.flush()
        writer.push(event)
        writer.push(event)

    test_file.seek(0)

    with proio.Reader(fileobj = test_file) as reader:
        n_events = 4
        reader.skip(n_events);
        for event in reader:
            n_events += 1
        assert(n_events == 5)

    test_file.close()

def test_write_seek_fifo():
    tmpdir = tempfile.mkdtemp()
    filename = os.path.join(tmpdir, 'test_fifo')
    os.mkfifo(filename)
    print("FIFO:", filename)

    threading.Thread(target=write_fifo, args=(filename,))
    threading.Thread(target=read_fifo, args=(filename,))

    os.remove(filename)
    os.rmdir(tmpdir)

def write_fifo(filename):
    with proio.Writer(filename) as writer:
        event = proio.Event()
        writer.push(event)
        writer.flush()
        writer.push(event)
        writer.push(event)
        writer.flush()
        writer.push(event)
        writer.push(event)

def read_fifo(filename):
    n_events = 4
    with proio.Reader(filename) as reader:
        reader.skip(n_events);
        for event in reader:
            n_events += 1
    assert(n_events == 5)

