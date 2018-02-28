import gzip
import io
import lz4.frame
import struct
import sys

from .event import Event
import proio.proto as proto
from .writer import magic_bytes

class Reader(object):
    """
    Reader for proio files
    
    This class can be used with the `with` statement, and it also may be used
    as an iterator that sequentially iterates all events.  A filename may be
    omitted in favor of specifying `fileobj`.

    :param string filename: name of input file to read
    :param fileobj: file object to read from

    :example:
    
    .. code-block:: python

        with proio.Reader('input.proio') as reader:
            for event in reader:
                ...

    """

    def __init__(self, filename = None, fileobj = None):
        if filename is None:
            if fileobj is not None:
                self._stream_reader = fileobj
            else:
                self._stream_reader = io.BytesIO(b'')
        else:
            self._stream_reader = open(filename, 'rb')
            self._close_file = True

        self._bucket_reader = io.BytesIO(b'')

    def __enter__(self):
        return self

    def __exit__(self, exception_type, exception_value, traceback):
        self.close()

    def close(self):
        """
        closes the underlying input file object.
        """
        try:
            if self._close_file:
                self._stream_reader.close()
        except:
            pass

    def next(self):
        """
        :return: the next event
        :rtype: Event
        """
        event = self._read_from_bucket(True)
        if event is None:
            raise StopIteration
        return event

    def next_header(self):
        """
        returns the next event header.  This is useful for scanning the
        stream/file.

        :return: the next event header
        :rtype: Event
        """
        self._read_bucket(sys.maxsize)
        return self._bucket_header

    def skip(self, n_events):
        """
        skips the next `n_events` events.

        :param int n_events: number of events to skip
        :return: number of events skipped
        :rtype: int
        """
        try:
            bucket_evts_left = self._bucket_header.nEvents - self._bucket_evts_read
        except AttributeError:
            bucket_evts_left = 0

        n_skipped = 0
        if n_events > bucket_evts_left:
            n_skipped += bucket_evts_left
            while True:
                n = self._read_bucket(n_events - n_skipped)
                if n == 0:
                    break
                n_skipped += n

        while n_skipped < n_events:
            if self._read_from_bucket(False) == True:
                n_skipped += 1
            else:
                break

        return n_skipped

    def seek_to_start(self):
        """
        seeks, if possible, to the start of the input file object.  This can be
        used along with :func:`skip` to directly access events.
        """
        if self._stream_reader.seekable():
            self._stream_reader.seek(0, 0)
            self._bucket_reader = io.BytesIO(b'')

    def _read_from_bucket(self, do_unmarshal = True):
        proto_size_buf = self._bucket_reader.read(4)
        if len(proto_size_buf) != 4:
            self._read_bucket()
            proto_size_buf = self._bucket_reader.read(4)
            if len(proto_size_buf) != 4:
                return

        proto_size = struct.unpack("I", proto_size_buf)[0]
        proto_buf = self._bucket_reader.read(proto_size)
        if len(proto_buf) != proto_size:
            return
        self._bucket_evts_read += 1

        if do_unmarshal:
            event_proto = proto.Event.FromString(proto_buf)
            return Event(proto_obj = event_proto)

        return True

    def _read_bucket(self, max_skip_events = 0):
        self._bucket_evts_read = 0
        events_skipped = 0
        self._bucket_header = None
        
        n = self._sync_to_magic()
        if n < len(magic_bytes):
            return events_skipped

        header_size = struct.unpack("I", self._stream_reader.read(4))[0]
        header_string = self._stream_reader.read(header_size)
        if len(header_string) != header_size:
            return events_skipped
        self._bucket_header = proto.BucketHeader.FromString(header_string)

        if self._bucket_header.nEvents > max_skip_events:
            bucket = self._stream_reader.read(self._bucket_header.bucketSize)
        else:
            self._bucket_reader = io.BytesIO(b'')
            events_skipped = self._bucket_header.nEvents
            try:
                self._stream_reader.seek(self._bucket_header.bucketSize, 1)
            except OSError:
                self._stream_reader.read(self._bucket_header.bucketSize)
            return events_skipped

        if len(bucket) != self._bucket_header.bucketSize:
            return events_skipped

        if self._bucket_header.compression == proto.BucketHeader.GZIP:
            self._bucket_reader = gzip.GzipFile(fileobj = io.BytesIO(bucket), mode = 'rb')
        elif self._bucket_header.compression == proto.BucketHeader.LZ4:
            try:
                uncomp_bytes, _ = lz4.frame.decompress(bucket)
            except ValueError:
                uncomp_bytes = lz4.frame.decompress(bucket)
            self._bucket_reader = io.BytesIO(uncomp_bytes)
        else:
            self._bucket_reader = io.BytesIO(bucket)

        return events_skipped

    def _sync_to_magic(self):
        n_read = 0
        while True:
            magic_byte = self._stream_reader.read(1)
            if len(magic_byte) != 1:
                return -1
            n_read += 1

            if magic_byte == magic_bytes[0]:
                goodSeq = True
                for i in range(1, len(magic_bytes)):
                    magic_byte = self._stream_reader.read(1)
                    if len(magic_byte) != 1:
                        return -1
                    n_read += 1

                    if magic_byte != magic_bytes[i]:
                        goodSeq = False
                        break
                if goodSeq:
                    break

        return n_read

    def __iter__(self):
        return self

    def __next__(self):
        return self.next()
