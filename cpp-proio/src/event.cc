#include <stdarg.h>
#include <algorithm>
#include <sstream>

#include "event.h"
#include "reader.h"

using namespace proio;
using namespace google::protobuf;

Event::Event(proto::Event *eventProto) {
    if (!eventProto)
        this->eventProto = new proto::Event();
    else
        this->eventProto = eventProto;
}

Event::~Event() {
    delete eventProto;
    for (auto idEntryPair : entryCache) {
        int64 id = idEntryPair.first;
        delete idEntryPair.second;
    }
}

uint64_t Event::AddEntry(std::string tag, Message *entry) {
    uint64_t typeID = getTypeID(entry);
    proto::Any entryProto;
    entryProto.set_type(typeID);

    eventProto->set_nentries(eventProto->nentries() + 1);
    uint64_t id = eventProto->nentries();
    (*eventProto->mutable_entries())[id] = entryProto;

    entryCache[id] = entry;

    TagEntry(id, tag);

    return id;
}

Message *Event::GetEntry(uint64_t id) {
    if (entryCache.count(id)) return entryCache[id];

    if (!eventProto->entries().count(id)) return NULL;
    const proto::Any entryProto = eventProto->entries().at(id);

    const std::string typeName = eventProto->types().at(entryProto.type());
    const Descriptor *desc = DescriptorPool::generated_pool()->FindMessageTypeByName(typeName);
    if (!desc) throw unknownMessageTypeError;
    Message *entry = MessageFactory::generated_factory()->GetPrototype(desc)->New();
    if (!entry->ParseFromString(entryProto.payload())) {
        delete entry;
        throw deserializationError;
    }
    entryCache[id] = entry;

    return entry;
}

void Event::TagEntry(uint64_t id, std::string tag) { (*eventProto->mutable_tags())[tag].add_entries(id); }

std::vector<std::string> Event::Tags() {
    std::vector<std::string> tags;
    for (auto stringTagPair : eventProto->tags()) {
        tags.push_back(stringTagPair.first);
    }
    std::sort(tags.begin(), tags.end());
    return tags;
}

std::vector<uint64_t> Event::TaggedEntries(std::string tag) {
    if (eventProto->tags().count(tag)) {
        auto entries = eventProto->tags().at(tag).entries();
        std::vector<uint64_t> returnEntries;
        for (uint64_t entry : entries) returnEntries.push_back(entry);
        return returnEntries;
    }
    return std::vector<uint64_t>();
}

std::string Event::String() {
    std::string printString;
    for (auto tag : Tags()) {
        printString += "---------- TAG: " + tag + " ----------\n";
        auto entries = TaggedEntries(tag);
        for (uint64_t entryID : entries) {
            std::stringstream ss;
            ss << "ID: " << entryID << "\n";
            Message *entry = GetEntry(entryID);
            if (entry) {
                ss << "Entry type: " << entry->GetTypeName() << "\n";
                ss << entry->DebugString() << "\n";
            } else
                ss << "not found\n";
            printString += ss.str();
        }
    }
    return printString;
}

void Event::flushCollCache() {
    for (auto idEntryPair : entryCache) {
        int64 id = idEntryPair.first;
        Message *entry = idEntryPair.second;

#if GOOGLE_PROTOBUF_VERSION >= 3004000
        size_t byteSize = entry->ByteSizeLong();
#else
        size_t byteSize = entry->ByteSize();
#endif
        uint8_t *buffer = new uint8_t[byteSize];
        entry->SerializeToArray(buffer, byteSize);
        delete entry;

        (*eventProto->mutable_entries())[id].set_payload(buffer, byteSize);
        delete[] buffer;
    }
    entryCache.clear();
}

proto::Event *Event::getProto() { return eventProto; }

uint64_t Event::getTypeID(Message *entry) {
    std::string typeName = entry->GetTypeName();
    if (revTypeLookup.count(typeName)) {
        return revTypeLookup[typeName];
    }

    for (auto typePair : eventProto->types()) {
        if (typePair.second.compare(typeName) == 0) {
            revTypeLookup[typeName] = typePair.first;
            return typePair.first;
        }
    }

    eventProto->set_ntypes(eventProto->ntypes() + 1);
    uint64_t typeID = eventProto->ntypes();
    (*eventProto->mutable_types())[typeID] = typeName;
    revTypeLookup[typeName] = typeID;
    return typeID;
}
