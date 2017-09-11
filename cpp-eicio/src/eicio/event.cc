#include <iostream>

#include "event.h"

using namespace google::protobuf;

eicio::Event::Event() { header = new eicio::EventHeader(); }

eicio::Event::~Event() {
    if (header) delete header;
    for (auto collEntry = collCache.begin(); collEntry != collCache.end(); collEntry++) {
        delete collEntry->second;
    }
}

Message *eicio::Event::Get(std::string collName) {
    Message *msg;
    if ((msg = collCache[collName]) != NULL) return msg;

    return getFromPayload(collName);
}

std::vector<std::string> eicio::Event::GetNames() {
    std::vector<std::string> names;

    for (int i = 0; i < header->payloadcollections_size(); i++) {
        auto collHdr = header->payloadcollections()[i];
        names.push_back(collHdr.name());
    }
    for (auto collEntry = collCache.begin(); collEntry != collCache.end(); collEntry++) {
        names.push_back(collEntry->first);
    }

    return names;
}

void eicio::Event::SetHeader(eicio::EventHeader *newHeader) {
    if (header) delete header;
    header = newHeader;
}

eicio::EventHeader *eicio::Event::GetHeader() { return header; }

unsigned int eicio::Event::GetPayloadSize() { return payload.size(); }

void *eicio::Event::SetPayloadSize(uint32 size) {
    payload.resize(size);
    return &payload[0];
}

unsigned char *eicio::Event::GetPayload() { return &payload[0]; }

std::string eicio::Event::GetType(Message *coll) {
    static const std::string prefix = "eicio.";
    return coll->GetTypeName().substr(prefix.length());
}

void eicio::Event::FlushCollCache() {
    for (int i = 0; i < namesCached.size(); i++) {
        auto name = namesCached[i];
        auto coll = collCache[name];
        collToPayload(coll, name);
        collCache.erase(name);
        namesCached.erase(namesCached.begin() + i);
    }
}

Message *eicio::Event::getFromPayload(std::string name, bool parse) {
    uint32 offset = 0;
    uint32 size = 0;
    std::string collType = "";
    int collIndex = 0;
    for (int i = 0; i < header->payloadcollections_size(); i++) {
        auto collHdr = header->payloadcollections()[i];
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
        auto desc = DescriptorPool::generated_pool()->FindMessageTypeByName("eicio." + collType);
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

void eicio::Event::collToPayload(Message *coll, std::string name) {
    const Descriptor *desc = coll->GetDescriptor();
    const Reflection *ref = coll->GetReflection();

    const FieldDescriptor *idFieldDesc = desc->FindFieldByName("id");
    if (!idFieldDesc) return;

    eicio::EventHeader_CollectionHeader *collHdr = header->add_payloadcollections();
    collHdr->set_name(name);
    collHdr->set_id(ref->GetUInt32(*coll, idFieldDesc));
    collHdr->set_type(GetType(coll));

    const size_t byteSize = coll->ByteSizeLong();
    size_t offset = payload.size();
    payload.resize(offset + byteSize);
    unsigned char *buf = &payload[0] + offset;
    coll->SerializeToArray(buf, byteSize);
}
