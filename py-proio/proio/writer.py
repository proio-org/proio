import gzip
import io
import lz4.frame
import struct

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

class Writer(object):
    """
    Writer for proio files
    
    This class can be used with the `with` statement.  A filename may be
    omitted in favor of specifying `fileobj`.

    :param string filename: name of output file to create or overwrite
    :param fileobj: file object to write to

    :example:
    
    .. code-block:: python

        with proio.Writer('output.proio') as writer:
            ...

    """

    def __init__(self, filename = None, fileobj = None):
        if filename is None:
            if fileobj is not None:
                self._stream_writer = fileobj
            else:
                self._stream_writer = io.BytesIO(b'')
        else:
            self._stream_writer = open(filename, 'wb')
            self._close_file = True

        self.bucket_dump_size = 0x1000000

        self._bucket_events = 0
        self._bucket = io.BytesIO(b'')
        self.set_compression(proto.BucketHeader.LZ4)

    def __enter__(self):
        return self

    def __exit__(self, exception_type, exception_value, traceback):
        self.close()

    def close(self):
        """
        closes the file object assigned to the Writer.  This is automatically
        called at the end of a `with` statement.
        """
        self.flush()
        try:
            if self._close_file:
                self._stream_writer.close()
        except:
            pass

    def flush(self):
        """
        flushes all buffered data to the output file object.  This is
        automatically called at the end of a `with` statement.
        """
        if self._bucket_events == 0:
            return

        if self._comp == proto.BucketHeader.LZ4:
            bucket_bytes = lz4.frame.compress(self._bucket.getvalue())
        elif self._comp == proto.BucketHeader.GZIP:
            bucket_compressed = io.BytesIO(b'')
            with gzip.GzipFile(fileobj = bucket_compressed, mode = 'wb') as writer:
                writer.write(self._bucket.getvalue())
            bucket_bytes = bucket_compressed.getvalue()
        else:
            bucket_bytes = self._bucket.getvalue()

        self._bucket.seek(0, 0)
        self._bucket.truncate(0)

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
        """
        sets the compression type to use for future output buckets.

        :param comp: can be one of :attr:`proio.LZ4`, :attr:`proio.GZIP`, or
            :attr:`proio.UNCOMPRESSED`
        """
        self._comp = comp

    def push(self, event):
        """
        takes an event and serializes it into the output bucket.

        :param Event event: event to serialize to output
        """
        event._flush_cache()
        proto_buf = event._proto.SerializeToString()

        proto_size = struct.pack("I", len(proto_buf))

        self._bucket.write(proto_size)
        self._bucket.write(proto_buf)

        self._bucket_events += 1

        bucket_length = len(self._bucket.getvalue())
        if bucket_length > self.bucket_dump_size:
            self.flush()
