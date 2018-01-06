#ifndef PROIO_EVENT_H
#define PROIO_EVENT_H

#include <string>

#include "proio.pb.h"

namespace proio {
class Event {
   public:
    Event(proto::Event *eventProto = NULL);
    virtual ~Event();

    uint64_t AddEntry(std::string tag, google::protobuf::Message *entry);
    google::protobuf::Message *GetEntry(uint64_t id);
    void TagEntry(uint64_t id, std::string tag);
    std::vector<std::string> Tags();
    google::protobuf::RepeatedField<uint64_t> TaggedEntries(std::string tag);

    std::string String();

    void flushCollCache();
    proto::Event *getProto();

   private:
    uint64_t getTypeID(google::protobuf::Message *entry);

    proto::Event *eventProto;
    std::map<std::string, uint64_t> revTypeLookup;
    std::map<uint64_t, google::protobuf::Message *> entryCache;
};

const class UnknownMessageTypeError : public std::exception {
    virtual const char *what() const throw() { return "Unknown message type"; }
} unknownMessageTypeError;
}  // namespace proio

#endif  // PROIO_EVENT_H
