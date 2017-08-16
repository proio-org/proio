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

   private:
    std::vector<unsigned char> payload;

    std::map<std::string, google::protobuf::Message*> collCache;
};
}

#endif  // EVENT_H
