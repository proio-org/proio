#ifndef EVENT_H
#define EVENT_H

#include "proio.pb.h"

namespace proio {
class Event {
   public:
    Event(proto::Event *eventProto = NULL);
    virtual ~Event();

    uint64_t AddEntry(std::string tag, google::protobuf::Message *entry);
    void TagEntry(uint64_t id, std::string tag);

    void flushCollCache();
    proto::Event *getProto();

   private:
    google::protobuf::uint64 getTypeID(google::protobuf::Message *entry);

    proto::Event *eventProto;
    std::map<std::string, google::protobuf::uint64> revTypeLookup;
    std::map<google::protobuf::uint64, google::protobuf::Message *> entryCache;
};
}  // namespace proio

#endif  // EVENT_H
