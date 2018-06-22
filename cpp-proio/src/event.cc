#include <stdarg.h>
#include <algorithm>
#include <sstream>

#include "event.h"
#include "reader.h"

using namespace proio;
using namespace google::protobuf;

Event::Event() {
    eventProto = new proto::Event();
    dirtyTags = false;
}

Event::Event(const std::string &data) : Event() { eventProto->ParseFromString(data); }

Event::~Event() {
    delete eventProto;
    for (auto idEntryPair : entryCache) delete idEntryPair.second;
    for (auto descVectorPair : store)
        for (auto entry : descVectorPair.second) delete entry;
}

uint64_t Event::AddEntry(Message *entry, std::string tag) {
    uint64_t typeID = getTypeID(entry);
    proto::Any entryProto;
    entryProto.set_type(typeID);

    eventProto->set_nentries(eventProto->nentries() + 1);
    uint64_t id = eventProto->nentries();
    (*eventProto->mutable_entry())[id] = entryProto;

    entryCache[id] = entry;

    if (tag.size() > 0) TagEntry(id, tag);

    return id;
}

Message *Event::GetEntry(uint64_t id) {
    if (entryCache.count(id)) return entryCache[id];

    if (!eventProto->entry().count(id)) return NULL;
    const proto::Any entryProto = eventProto->entry().at(id);

    const Descriptor *desc = getDescriptor(entryProto.type());
    if (!desc) throw unknownMessageTypeError;
    Message *entry;
    std::vector<Message *> &storeEntries = store[desc];
    if (storeEntries.size() > 0) {
        entry = storeEntries.back();
        storeEntries.pop_back();
    } else
        entry = MessageFactory::generated_factory()->GetPrototype(desc)->New();
    if (!entry->ParseFromString(entryProto.payload())) {
        entry->Clear();
        store[entry->GetDescriptor()].push_back(entry);
        throw deserializationError;
    }
    entryCache[id] = entry;

    return entry;
}

void Event::TagEntry(uint64_t id, std::string tag) { (*eventProto->mutable_tag())[tag].add_entry(id); }

void Event::UntagEntry(uint64_t id, std::string tag) {
    if (!eventProto->tag().count(tag)) return;

    auto entries = eventProto->mutable_tag()->at(tag).mutable_entry();
    for (auto iter = entries->begin(); iter != entries->end(); iter++) {
        if ((*iter) == id) {
            entries->erase(iter);
            break;
        }
    }
}

void Event::RemoveEntry(uint64_t id) {
    if (entryCache.count(id)) {
        Message *entry = entryCache[id];
        entryCache.erase(id);
        entry->Clear();
        store[entry->GetDescriptor()].push_back(entry);
    }
    eventProto->mutable_entry()->erase(id);
    dirtyTags = true;
}

std::vector<std::string> Event::Tags() {
    std::vector<std::string> tags;
    for (auto stringTagPair : eventProto->tag()) {
        tags.push_back(stringTagPair.first);
    }
    std::sort(tags.begin(), tags.end());
    return tags;
}

std::vector<uint64_t> Event::TaggedEntries(std::string tag) {
    if (eventProto->tag().count(tag)) {
        tagCleanup();
        auto entries = eventProto->tag().at(tag).entry();
        std::vector<uint64_t> returnEntries;
        for (uint64_t entry : entries) returnEntries.push_back(entry);
        return returnEntries;
    }
    return std::vector<uint64_t>();
}

std::vector<uint64_t> Event::AllEntries() {
    auto entries = eventProto->entry();
    std::vector<uint64_t> returnEntries;
    for (auto idEntryPair : entries) returnEntries.push_back(idEntryPair.first);
    return returnEntries;
}

std::vector<std::string> Event::EntryTags(uint64_t id) {
    std::vector<std::string> tags;
    for (auto stringTagPair : eventProto->tag()) {
        for (uint64_t entry : stringTagPair.second.entry())
            if (entry == id) {
                tags.push_back(stringTagPair.first);
                break;
            }
    }
    std::sort(tags.begin(), tags.end());
    return tags;
}

void Event::DeleteTag(std::string tag) { eventProto->mutable_tag()->erase(tag); }

Message *Event::Free(const Descriptor *desc) {
    std::vector<Message *> &storeEntries = store[desc];
    if (storeEntries.size() > 0) {
        Message *entry = storeEntries.back();
        storeEntries.pop_back();
        return entry;
    } else
        return NULL;
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

void Event::FlushCache() {
    for (auto idEntryPair : entryCache) {
        int64 id = idEntryPair.first;
        Message *entry = idEntryPair.second;

        size_t byteSize = entry->ByteSizeLong();
        uint8_t *buffer = new uint8_t[byteSize];
        entry->SerializeToArray(buffer, byteSize);
        entry->Clear();
        store[entry->GetDescriptor()].push_back(entry);

        (*eventProto->mutable_entry())[id].set_payload(buffer, byteSize);
        delete[] buffer;
    }
    entryCache.clear();

    tagCleanup();
}

void Event::Clear() {
    eventProto->Clear();
    revTypeLookup.clear();
    for (auto idEntryPair : entryCache) {
        Message *entry = idEntryPair.second;
        entry->Clear();
        store[entry->GetDescriptor()].push_back(entry);
    }
    entryCache.clear();
    descriptorCache.clear();
    metadata.clear();
    dirtyTags = false;
}

Event &Event::operator=(const Event &event) {
    if (&event == this) return *this;
    *this->eventProto = *event.eventProto;
    this->revTypeLookup = event.revTypeLookup;
    for (auto idEntryPair : event.entryCache) {
        auto entry = idEntryPair.second;
        const Descriptor *desc = getDescriptor(getTypeID(entry));
        if (!desc) throw unknownMessageTypeError;
        auto newEntry = MessageFactory::generated_factory()->GetPrototype(desc)->New();
        newEntry->MergeFrom(*entry);
        this->entryCache[idEntryPair.first] = newEntry;
    }
    this->descriptorCache = event.descriptorCache;
    this->metadata = event.metadata;
    this->dirtyTags = event.dirtyTags;
    return *this;
}

proto::Event *Event::getProto() { return eventProto; }

uint64_t Event::getTypeID(Message *entry) {
    std::string typeName = entry->GetTypeName();
    if (revTypeLookup.count(typeName)) {
        return revTypeLookup[typeName];
    }

    for (auto typePair : eventProto->type()) {
        if (typePair.second.compare(typeName) == 0) {
            revTypeLookup[typeName] = typePair.first;
            return typePair.first;
        }
    }

    eventProto->set_ntypes(eventProto->ntypes() + 1);
    uint64_t typeID = eventProto->ntypes();
    (*eventProto->mutable_type())[typeID] = typeName;
    revTypeLookup[typeName] = typeID;
    return typeID;
}

const Descriptor *Event::getDescriptor(uint64_t typeID) {
    if (!descriptorCache.count(typeID)) {
        const std::string typeName = eventProto->type().at(typeID);
        descriptorCache[typeID] = DescriptorPool::generated_pool()->FindMessageTypeByName(typeName);
    }
    return descriptorCache[typeID];
}

void Event::tagCleanup() {
    if (!dirtyTags) return;
    auto tags = eventProto->mutable_tag();
    for (auto iter = tags->begin(); iter != tags->end(); iter++) {
        RepeatedField<uint64_t> *entryList = iter->second.mutable_entry();
        for (int i = entryList->size() - 1; i >= 0; i--) {
            if (!eventProto->entry().count((*entryList)[i])) {
                for (int j = i; j < entryList->size() - 1; j++) entryList->Set(j, entryList->Get(j + 1));
                entryList->RemoveLast();
            }
        }
    }
    dirtyTags = false;
}
