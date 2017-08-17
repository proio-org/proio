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

void *eicio::Event::SetPayloadSize(uint32 size) {
    payload.resize(size);
    return &payload[0];
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
    }

    header->mutable_payloadcollections()->DeleteSubrange(collIndex, 1);
    payload.erase(payload.begin() + offset, payload.begin() + offset + size);

    return coll;
}
