#include <iostream>

#include "event.h"

using namespace google::protobuf;

using namespace eicio;

Event::Event() { header = new model::EventHeader(); }

Event::~Event() {
    if (header) delete header;
    for (auto collEntry = collCache.begin(); collEntry != collCache.end(); collEntry++) {
        delete collEntry->second;
    }
}

bool Event::Add(Message *coll, std::string name) {
    const Descriptor *desc = coll->GetDescriptor();
    const Reflection *refl = coll->GetReflection();
    const FieldDescriptor *idFieldDesc = desc->FindFieldByName("id");
    if (!idFieldDesc) return false;
    uint32 collID = refl->GetUInt32(*coll, idFieldDesc);

    for (auto collPair : collCache) {
        if (collPair.first.compare(name) == 0) return false;

        auto collEx = collPair.second;

        desc = collEx->GetDescriptor();
        refl = collEx->GetReflection();
        idFieldDesc = desc->FindFieldByName("id");
        if (!idFieldDesc) continue;
        uint32 collIDEx = refl->GetUInt32(*collEx, idFieldDesc);

        if (collIDEx == collID) return false;
    }

    for (auto collHdr : header->payloadcollections()) {
        if (collHdr.name().compare(name) == 0) return false;
        if (collHdr.id() == collID) return false;
    }

    collCache[name] = coll;
    namesCached.push_back(name);
    return true;
}

void Event::Remove(std::string collName) {
    if (collCache.find(collName) != collCache.end()) {
        delete collCache[collName];
        collCache.erase(collName);
        return;
    }

    for (auto collHdr : header->payloadcollections()) {
        if (collHdr.name().compare(collName) == 0) {
            delete getFromPayload(collName);
            return;
        }
    }
}

Message *Event::Get(std::string collName) {
    Message *msg;
    if ((msg = collCache[collName]) != NULL) return msg;

    return getFromPayload(collName);
}

std::vector<std::string> Event::GetNames() {
    std::vector<std::string> names;

    for (int i = 0; i < header->payloadcollections_size(); i++) {
        auto collHdr = header->payloadcollections(i);
        names.push_back(collHdr.name());
    }
    for (auto collEntry = collCache.begin(); collEntry != collCache.end(); collEntry++) {
        names.push_back(collEntry->first);
    }

    return names;
}

void Event::Reference(Message *msg, model::Reference *ref) {
    for (auto collPair : collCache) {
        auto coll = collPair.second;

        const Descriptor *desc = coll->GetDescriptor();
        const Reflection *refl = coll->GetReflection();

        const FieldDescriptor *idFieldDesc = desc->FindFieldByName("id");
        if (!idFieldDesc) continue;
        uint32 collID = refl->GetUInt32(*coll, idFieldDesc);

        if (coll == msg) {
            if (collID == 0) {
                collID = GetUniqueID();
                refl->SetUInt32(coll, idFieldDesc, collID);
            }

            ref->set_collid(collID);
            return;
        }

        const FieldDescriptor *entriesFieldDesc = desc->FindFieldByName("entries");
        if (!entriesFieldDesc) continue;
        RepeatedPtrField<Message> *entries = refl->MutableRepeatedPtrField<Message>(coll, entriesFieldDesc);

        for (int i = 0; i < entries->size(); i++) {
            if (&(*entries)[i] == msg) {
                const Descriptor *entryDesc = (*entries)[i].GetDescriptor();
                const Reflection *entryRefl = (*entries)[i].GetReflection();

                const FieldDescriptor *entryIDFieldDesc = entryDesc->FindFieldByName("id");
                if (!entryIDFieldDesc) continue;
                unsigned int entryID = entryRefl->GetUInt32((*entries)[i], entryIDFieldDesc);

                if (collID == 0) {
                    collID = GetUniqueID();
                    refl->SetUInt32(coll, idFieldDesc, collID);
                }
                if (entryID == 0) {
                    entryID = GetUniqueID();
                    entryRefl->SetUInt32(&(*entries)[i], entryIDFieldDesc, entryID);
                }

                ref->set_collid(collID);
                ref->set_entryid(entryID);
                return;
            }
        }
    }

    return;
}

Message *Event::Dereference(const model::Reference &ref) {
    Message *refColl = NULL;
    for (auto collPair : collCache) {
        auto coll = collPair.second;

        const Descriptor *desc = coll->GetDescriptor();
        const Reflection *refl = coll->GetReflection();

        const FieldDescriptor *idFieldDesc = desc->FindFieldByName("id");
        if (!idFieldDesc) continue;
        uint32 collID = refl->GetUInt32(*coll, idFieldDesc);

        if (collID == ref.collid()) {
            refColl = coll;
            if (ref.entryid() == 0) return refColl;
            break;
        }
    }
    if (refColl == NULL) {
        for (auto collHdr : header->payloadcollections()) {
            if (collHdr.id() == ref.collid()) {
                refColl = Get(collHdr.name());
                if (ref.entryid() == 0) return refColl;
                break;
            }
        }
    }
    if (refColl == NULL) return NULL;

    const Descriptor *desc = refColl->GetDescriptor();
    const Reflection *refl = refColl->GetReflection();

    const FieldDescriptor *entriesFieldDesc = desc->FindFieldByName("entries");
    if (!entriesFieldDesc) return NULL;
    RepeatedPtrField<Message> *entries = refl->MutableRepeatedPtrField<Message>(refColl, entriesFieldDesc);

    for (int i = 0; i < entries->size(); i++) {
        const Descriptor *entryDesc = (*entries)[i].GetDescriptor();
        const Reflection *entryRefl = (*entries)[i].GetReflection();

        const FieldDescriptor *entryIDFieldDesc = entryDesc->FindFieldByName("id");
        if (!entryIDFieldDesc) continue;
        unsigned int entryID = entryRefl->GetUInt32((*entries)[i], entryIDFieldDesc);

        if (entryID == ref.entryid()) {
            return &(*entries)[i];
        }
    }

    return NULL;
}

unsigned int Event::GetUniqueID() {
    header->set_nuniqueids(header->nuniqueids() + 1);
    return header->nuniqueids();
}

void Event::SetHeader(model::EventHeader *newHeader) {
    if (header) delete header;
    header = newHeader;
}

model::EventHeader *Event::GetHeader() { return header; }

unsigned int Event::GetPayloadSize() { return payload.size(); }

void *Event::SetPayloadSize(uint32 size) {
    payload.resize(size);
    return &payload[0];
}

unsigned char *Event::GetPayload() { return &payload[0]; }

std::string Event::GetType(Message *coll) {
    static const std::string prefix = "eicio.model.";
    return coll->GetTypeName().substr(prefix.length());
}

void Event::FlushCollCache() {
    for (int i = 0; i < namesCached.size(); i++) {
        auto name = namesCached[i];
        auto coll = collCache[name];
        collToPayload(coll, name);
        collCache.erase(name);
        namesCached.erase(namesCached.begin() + i);
        delete coll;
    }
}

Message *Event::getFromPayload(std::string name, bool parse) {
    uint32 offset = 0;
    uint32 size = 0;
    std::string collType = "";
    int collIndex = 0;
    for (int i = 0; i < header->payloadcollections_size(); i++) {
        auto collHdr = header->payloadcollections(i);
        if (name.compare(collHdr.name()) == 0) {
            collType = collHdr.type();
            size = collHdr.payloadsize();
            break;
        }
        offset += collHdr.payloadsize();
    }
    if (collType.length() == 0) {
        return NULL;
    }

    Message *coll;
    if (parse) {
        auto desc = DescriptorPool::generated_pool()->FindMessageTypeByName("eicio.model." + collType);
        if (desc == NULL) {
            return NULL;
        }
        coll = MessageFactory::generated_factory()->GetPrototype(desc)->New();

        if (!coll->ParseFromArray(&payload[0] + offset, size)) {
            delete coll;
            return NULL;
        }

        collCache[name] = coll;
        namesCached.push_back(name);
    }

    header->mutable_payloadcollections()->DeleteSubrange(collIndex, 1);
    payload.erase(payload.begin() + offset, payload.begin() + offset + size);

    return coll;
}

void Event::collToPayload(Message *coll, std::string name) {
    const Descriptor *desc = coll->GetDescriptor();
    const Reflection *ref = coll->GetReflection();

    const FieldDescriptor *idFieldDesc = desc->FindFieldByName("id");
    if (!idFieldDesc) return;

    model::EventHeader_CollectionHeader *collHdr = header->add_payloadcollections();
    collHdr->set_name(name);
    collHdr->set_id(ref->GetUInt32(*coll, idFieldDesc));
    collHdr->set_type(GetType(coll));

    const size_t byteSize = coll->ByteSizeLong();
    collHdr->set_payloadsize(byteSize);
    size_t offset = payload.size();
    payload.resize(offset + byteSize);
    unsigned char *buf = &payload[0] + offset;
    coll->SerializeToArray(buf, byteSize);
}
