import gzip
import struct

from .event import Event
import eicio.model as model
from .writer import magicBytes

class Reader:
    """Reader for eicio files, either gzip compressed or not"""

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
        try:
            self.fileobj.fileobj.close()
        except AttributeError:
            pass
        self.fileobj.close()

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        self.close()

    def get(self):
        n = self._syncToMagic()
        if n < 4:
            return
        
        sizes = struct.unpack("II", self.fileobj.read(8))
        headerSize = sizes[0]
        payloadSize = sizes[1]

        headerString = self.fileobj.read(headerSize)
        if len(headerString) != headerSize:
            return
        header = model.EventHeader.FromString(headerString)

        payload = self.fileobj.read(payloadSize)
        if len(payload) != payloadSize:
            return

        event = Event()
        event.header = header
        event._payload = payload

        return event

    def getHeader(self):
        n = self._syncToMagic()
        if n < 4:
            return
        
        sizes = struct.unpack("II", self.fileobj.read(8))
        headerSize = sizes[0]
        payloadSize = sizes[1]

        headerString = self.fileobj.read(headerSize)
        if len(headerString) != headerSize:
            return
        header = model.EventHeader.FromString(headerString)

        payload = self.fileobj.read(payloadSize)
        if len(payload) != payloadSize:
            return

        return header

    def _syncToMagic(self):
        nRead = 0
        while True:
            magicByte = self.fileobj.read(1)
            if len(magicByte) != 1:
                return -1
            nRead += 1

            if magicByte == magicBytes[0]:
                goodSeq = True
                for i in range(1, 4):
                    magicByte = self.fileobj.read(1)
                    if len(magicByte) != 1:
                        return -1
                    nRead += 1

                    if magicByte != magicBytes[i]:
                        goodSeq = False
                        break
                if goodSeq:
                    break

        return nRead

    def __iter__(self):
        return self

    def __next__(self):
        event = self.get()
        if event == None:
            raise StopIteration
        return event

    def seekToStart(self):
        if self.fileobj.seekable():
            self.fileobj.seek(0, 0)

    def skip(self, nEvents):
        nSkipped = 0
        for i in range(0, nEvents):
            n = self._syncToMagic()
            if n < 4:
                return
            
            sizes = struct.unpack("II", self.fileobj.read(8))
            size = sizes[0] + sizes[1]

            if self.fileobj.seekable():
                currPos = self.fileobj.seek(0, 1)
                postPos = self.fileobj.seek(size, 1)
                if postPos - currPos != size:
                    return nSkipped
            else:
                string = self.fileobj.read(size)
                if len(string) != size:
                    return nSkipped

            nSkipped += 1

        return nSkipped
