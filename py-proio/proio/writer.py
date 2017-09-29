import gzip
import struct

magic_bytes = [b'\xe1', b'\xc1', b'\x00', b'\x00']

class Writer:
    """Writer for proio files, either gzip compressed or not"""

    def __init__(self, filename = "", fileobj = None, compress = False):
        if fileobj != None:
            if compress:
                self.fileobj = gzip.GzipFile(fileobj = fileobj, mode = "wb")
            else:
                self.fileobj = fileobj
        else:
            if filename.endswith(".gz"):
                self.fileobj = gzip.open(filename, "wb")
            else:
                self.fileobj = open(filename, "wb")

    def close(self):
        self.fileobj.close()
        try:
            self.fileobj.fileobj.close()
        except AttributeError:
            pass

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()

    def push(self, event):
        event._flush_coll_cache()

        header_buf = event.header.SerializeToString()
        payload = event._payload
        meta_buf = struct.pack("II", len(header_buf), len(payload))

        for magic_byte in magic_bytes:
            self.fileobj.write(magic_byte)
        self.fileobj.write(meta_buf)
        self.fileobj.write(header_buf)
        self.fileobj.write(payload)

