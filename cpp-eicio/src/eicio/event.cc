#include <iostream>

#include "event.h"

using namespace google::protobuf;

eicio::Event::Event() { header = new eicio::EventHeader(); }

eicio::Event::~Event() {
    if (header) delete header;
}

Message *eicio::Event::Get(std::string collName) { ; }

void eicio::Event::SetHeader(eicio::EventHeader *newHeader) {
    if (header) delete header;
    header = newHeader;
}

eicio::EventHeader *eicio::Event::GetHeader() { return header; }

void *eicio::Event::SetPayloadSize(uint32 size) {
    payload.resize(size);
    return &payload[0];
}
