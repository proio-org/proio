import eicio.model as model

class Event:
    """Class representing a single event"""

    def get(self, collName):
        try:
            return self._collCache[collName]
        except AttributeError:
            self._collCache = {}
        except KeyError:
            pass

        return self._getFromPayload(collName)

    def getNames(self):
        names = []
        for collHdr in self.header.payloadCollections:
            names.append(collHdr.name)

        return names

    def _getFromPayload(self, collName, unmarshal = True):
        offset = 0
        size = 0
        collType = ""
        for collIndex in range(0, len(self.header.payloadCollections)):
            collHdr = self.header.payloadCollections[collIndex]
            if collHdr.name == collName:
                collType = collHdr.type
                size = collHdr.payloadSize
                break
            offset += collHdr.payloadSize
        if collType == "":
            return

        if unmarshal:
            messageClass = getattr(model, collType)
            message = messageClass.FromString(self._payload[offset : offset+size])
            self._collCache[collName] = message

        self.header.payloadCollections.remove(collHdr)
        self._payload = self._payload[:offset] + self._payload[offset+size:]

        return message
