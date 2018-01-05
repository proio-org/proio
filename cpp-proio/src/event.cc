#include "event.h"

using namespace proio;

Event::Event(proto::Event *eventProto) {
    if (!eventProto)
        this->eventProto = new proto::Event();
    else
        this->eventProto = eventProto;
}

Event::~Event() { delete eventProto; }

void Event::flushCollCache() { ; }

proto::Event *Event::getProto() { return eventProto; }
