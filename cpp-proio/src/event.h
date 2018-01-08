#ifndef PROIO_EVENT_H
#define PROIO_EVENT_H

#include <string>

#include "proio.pb.h"

namespace proio {
/** Class representing a single event
 */
class Event {
   public:
    Event(proto::Event *eventProto = NULL);
    virtual ~Event();

    /** AddEntry takes a tag and protobuf message entry and adds it to the
     * Event.  The return value is a uint64_t ID number used to reference the
     * added entry.
     */
    uint64_t AddEntry(std::string tag, google::protobuf::Message *entry);
    /** GetEntry takes an entry ID and returns the corresponding entry.
     */
    google::protobuf::Message *GetEntry(uint64_t id);
    /** TagEntry adds a tag to an entry that has already been added, identified
     * by its ID.
     */
    void TagEntry(uint64_t id, std::string tag);
    /** Tags returns a list of tags that exist in the event.
     */
    std::vector<std::string> Tags();
    /** TaggedEntries tages a tag string and returns a list of entry IDs that
     * the tag references.
     */
    const google::protobuf::RepeatedField<uint64_t> &TaggedEntries(std::string tag);

    /** String returns a human-readable string representing the event.
     */
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
