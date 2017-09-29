import gzip
import struct

from .event import Event
import proio.model as model
from .writer import magic_bytes

class Reader:
    """Reader for proio files, either gzip compressed or not"""

    def __init__(self, filename = "", fileobj = None, decompress = False):
        if fileobj != None:
            if decompress:
                self.fileobj = gzip.GzipFile(fileobj = fileobj, mode = "rb")
            else:
                self.fileobj = fileobj
        else:
            if filename.endswith(".gz"):
                self.fileobj = gzip.open(filename, "rb")
            else:
                self.fileobj = open(filename, "rb")

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

    def get(self):
        n = self._sync_to_magic()
        if n < 4:
            return
        
        sizes = struct.unpack("II", self.fileobj.read(8))
        header_size = sizes[0]
        payload_size = sizes[1]

        header_string = self.fileobj.read(header_size)
        if len(header_string) != header_size:
            return
        header = model.EventHeader.FromString(header_string)

        payload = self.fileobj.read(payload_size)
        if len(payload) != payload_size:
            return

        event = Event()
        event.header = header
        event._payload = payload

        return event

    def get_header(self):
        n = self._sync_to_magic()
        if n < 4:
            return
        
        sizes = struct.unpack("II", self.fileobj.read(8))
        header_size = sizes[0]
        payload_size = sizes[1]

        header_string = self.fileobj.read(header_size)
        if len(header_string) != header_size:
            return
        header = model.EventHeader.FromString(header_string)

        payload = self.fileobj.read(payload_size)
        if len(payload) != payload_size:
            return

        return header

    def _sync_to_magic(self):
        n_read = 0
        while True:
            magic_byte = self.fileobj.read(1)
            if len(magic_byte) != 1:
                return -1
            n_read += 1

            if magic_byte == magic_bytes[0]:
                goodSeq = True
                for i in range(1, 4):
                    magic_byte = self.fileobj.read(1)
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
        event = self.get()
        if event == None:
            raise StopIteration
        return event

    def seek_to_start(self):
        if self.fileobj.seekable():
            self.fileobj.seek(0, 0)

    def skip(self, n_events):
        n_skipped = 0
        for i in range(0, n_events):
            n = self._sync_to_magic()
            if n < 4:
                return
            
            sizes = struct.unpack("II", self.fileobj.read(8))
            size = sizes[0] + sizes[1]

            if self.fileobj.seekable():
                curr_pos = self.fileobj.seek(0, 1)
                post_pos = self.fileobj.seek(size, 1)
                if post_pos - curr_pos != size:
                    return n_skipped
            else:
                string = self.fileobj.read(size)
                if len(string) != size:
                    return n_skipped

            n_skipped += 1

        return n_skipped
