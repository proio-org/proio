import tempfile

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
