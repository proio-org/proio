import proio.proto as proto

import google.protobuf.descriptor_pool as descriptor_pool
import google.protobuf.message_factory as message_factory

class Event:
    """Class representing a single event"""

    def __init__(self, proto = None):
        self._proto = proto or proto.Event()
        self._entry_cache = {}
        self._factory = message_factory.MessageFactory()

    def get_entry(self, ID):
        try:
            return self._entry_cache[ID]
        except KeyError:
            pass

        try:
            entry_proto = self._proto.entries[ID]
        except KeyError:
            return

        type_string = self._proto.types[entry_proto.type]
        type_desc = descriptor_pool.Default().FindMessageTypeByName(type_string)
        msg_class = self._factory.GetPrototype(type_desc)
        entry = msg_class.FromString(entry_proto.payload)
        self._entry_cache[ID] = entry

        return entry

    def tags(self):
        tags = self._proto.tags.keys()
        tags.sort()
        return tags

    def tagged_entries(self, tag):
        return self._proto.tags[tag].entries
