import eicio.model as model

class Event:
    """Class representing a single event"""

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
            message_class = getattr(model, coll_type)
            message = message_class.FromString(self._payload[offset : offset+size])
            self._coll_cache[coll_name] = message

        self.header.payloadCollections.remove(coll_hdr)
        self._payload = self._payload[:offset] + self._payload[offset+size:]

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
