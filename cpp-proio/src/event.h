#ifndef EVENT_H
#define EVENT_H

#include "proio.pb.h"

namespace proio {
class Event {
   public:
    Event(proto::Event *eventProto = NULL);
    virtual ~Event();

    void flushCollCache();
    proto::Event *getProto();

   private:
    proto::Event *eventProto;
};
}  // namespace proio

#endif  // EVENT_H
