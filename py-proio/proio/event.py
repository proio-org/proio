import proio.model as model
import importlib

class Event:
    """Class representing a single event"""

    def __init__(self):
        self.header = model.EventHeader()
        self._payload = b''

    def get(self, coll_name):
        try:
            return self._coll_cache[coll_name]
        except AttributeError:
            self._coll_cache = {}
        except KeyError:
            pass

        return self._get_from_payload(coll_name)

    def get_names(self):
        names = []
        for coll_hdr in self.header.payloadCollections:
            names.append(coll_hdr.name)

        return names

    def add(self, coll, name):
        try:
            for key, coll_ in self._coll_cache.items():
                if key == name:
                    return
                if coll_.id != 0 and coll_.id == coll.id:
                    return
        except AttributeError:
            self._coll_cache = {}

        for coll_hdr in self.header.payloadCollections:
            if coll_hdr.name == name:
                return
            if coll_hdr.id != 0 and coll_hdr.id == coll.id:
                return

        self._coll_cache[name] = coll

    def remove(self, name):
        try:
            for key, _ in self._coll_cache.items():
                if key == name:
                    del self._coll_cache[name]
                    return
        except AttributeError:
            self._coll_cache = {}

        for coll_hdr in self.header.payloadCollections:
            if coll_hdr.name == name:
                self._get_from_payload(name, False)
                return

    def _get_from_payload(self, coll_name, unmarshal = True):
        offset = 0
        size = 0
        coll_type = ""
        for collIndex in range(0, len(self.header.payloadCollections)):
            coll_hdr = self.header.payloadCollections[collIndex]
            if coll_hdr.name == coll_name:
                coll_type = coll_hdr.type
                size = coll_hdr.payloadSize
                break
            offset += coll_hdr.payloadSize
        if coll_type == "":
            return

        if unmarshal:
            module_name = "proio.model"
            class_name = coll_type
            dot_i = coll_type.rfind(".")
            if dot_i > 0:
                module_name += "." + coll_type[:dot_i]
                class_name = coll_type[dot_i + 1:]
            message_module = importlib.import_module(module_name)
            message_class = getattr(message_module, class_name)
            message = message_class.FromString(self._payload[offset : offset+size])
            self._coll_cache[coll_name] = message

        self.header.payloadCollections.remove(coll_hdr)
        self._payload = self._payload[:offset] + self._payload[offset+size:]

        if unmarshal:
            return message

    def dereference(self, ref):
        refColl = None
        try:
            for name, coll in self._coll_cache.items():
                if coll.id == ref.collID:
                    if ref.entryID == 0:
                        return coll
                    refColl = coll
                    break
        except AttributeError:
            pass
        
        if refColl == None:
            for coll_hdr in self.header.payloadCollections:
                if coll_hdr.id == ref.collID:
                    refColl = self.get(coll_hdr.name)
                    if ref.entryID == 0:
                        return refColl
                    break

        if refColl == None:
            return

        for entry in refColl.entries:
            if entry.id == ref.entryID:
                return entry

        return

    def _flush_coll_cache(self):
        try:
            for name, coll in self._coll_cache.items():
                self._coll_to_payload(coll, name)
            self._coll_cache = {}
        except AttributeError:
            self._coll_cache = {}

    def _coll_to_payload(self, coll, name):
        if self.header == None:
            self.header = model.EventHeader()

        coll_hdr = self.header.payloadCollections.add()
        coll_hdr.name = name
        coll_hdr.id = coll.id
        coll_hdr.type = coll.DESCRIPTOR.full_name[12:]

        coll_buf = coll.SerializeToString()
        coll_hdr.payloadSize = len(coll_buf)

        self._payload = self._payload + coll_buf

    def get_unique_id(self):
        self.header.nUniqueIDs += 1
        return self.header.nUniqueIDs

    def reference(self, msg):
        try:
            for name, coll in self._coll_cache.items():
                if coll is msg:
                    coll_id = coll.id
                    if coll_id == 0:
                        coll_id = self.get_unique_id()
                        coll.id = coll_id
                    ref = model.Reference()
                    ref.collID = coll_id
                    ref.entryID = 0
                    return ref

                for entry in coll.entries:
                    if entry is msg:
                        coll_id = coll.id
                        if coll_id == 0:
                            coll_id = self.get_unique_id()
                            coll.id = coll_id
                        entry_id = entry.id
                        if entry_id == 0:
                            entry_id = self.get_unique_id()
                            entry.id = entry_id
                        ref = model.Reference()
                        ref.collID = coll_id
                        ref.entryID = entry_id
                        return ref
        except AttributeError:
            self._coll_cache = {}
