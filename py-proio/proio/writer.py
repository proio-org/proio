import gzip
import io
import lz4.frame
import struct

from .event import Event
import proio.proto as proto

magic_bytes = [b'\xe1',
        b'\xc1',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00',
        b'\x00']

class Writer:
    """Writer for proio files"""

    def __init__(self, filename = "", fileobj = None):
        if fileobj != None:
            self._stream_writer = fileobj
        else:
            self._stream_writer = open(filename, 'wb')

        self.bucket_dump_size = 0x1000000

        self._bucket_events = 0
        self._bucket = io.BytesIO(b'')
        self._bucket_writer = self._bucket
        self.set_compression(proto.BucketHeader.LZ4)

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()

    def close(self):
        self.flush()
        self._stream_writer.close()

    def flush(self):
        if self._bucket_events == 0:
            return

        self._bucket_writer.flush()
        if self._comp == proto.BucketHeader.LZ4:
            self._bucket.write(lz4.frame.compress(self._bucket_writer.read()))
        self._bucket.seek(0, 0)

        bucket_bytes = self._bucket.read()

        header = proto.BucketHeader()
        header.nEvents = self._bucket_events
        header.bucketSize = len(bucket_bytes)
        header.compression = self._comp
        header_buf = header.SerializeToString()

        header_size = struct.pack("I", len(header_buf))

        for magic_byte in magic_bytes:
            self._stream_writer.write(magic_byte)
        self._stream_writer.write(header_size)
        self._stream_writer.write(header_buf)
        self._stream_writer.write(bucket_bytes)

        self._bucket_events = 0

    def set_compression(self, comp):
        try:
            self.flush()
        except AttributeError:
            pass
        self._comp = comp

        if comp == proto.BucketHeader.GZIP:
            self._bucket_writer = gzip.GzipFile(fileobj = self._bucket, mode = 'wb')
        elif comp == proto.BucketHeader.LZ4:
            self._bucket_writer = io.BytesIO(b'')
        else:
            self._bucket_writer = self._bucket

    def push(self, event):
        event._flush_cache()
        proto_buf = event._proto.SerializeToString()

        proto_size = struct.pack("I", len(proto_buf))

        self._bucket_writer.write(proto_size)
        self._bucket_writer.write(proto_buf)

        self._bucket_events += 1

        bucket_length = len(self._bucket.getvalue())
        if self._comp == proto.BucketHeader.LZ4:
            bucket_length += len(self._bucket_writer.getvalue())

        if bucket_length > self.bucket_dump_size:
            self.flush()
