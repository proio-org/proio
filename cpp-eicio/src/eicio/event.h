#ifndef EVENT_H
#define EVENT_H

#include <map>
#include <vector>

#include "eicio.pb.h"

namespace eicio {
class Event {
   public:
    Event();
    virtual ~Event();

    google::protobuf::Message *Get(std::string collName);
    std::vector<std::string> GetNames();

    void SetHeader(EventHeader *newHeader);
    EventHeader *GetHeader();
    void *SetPayloadSize(google::protobuf::uint32 size);

   private:
    google::protobuf::Message *getFromPayload(std::string collName, bool parse = true);

    EventHeader *header;
    std::vector<unsigned char> payload;

    std::map<std::string, google::protobuf::Message *> collCache;
};
}

#endif  // EVENT_H
