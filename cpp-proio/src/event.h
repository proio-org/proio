#ifndef PROIO_EVENT_H
#define PROIO_EVENT_H

#include <string>

#include "proio.pb.h"

namespace proio {
/** Class representing a single event
 */
class Event {
   public:
    Event();
    /** This constructor initializes the event from the protobuf wire format.
     */
    Event(const std::string &proto);
    virtual ~Event();

    /** AddEntry takes a tag and protobuf message entry and adds it to the
     * Event.  The return value is a uint64_t ID number used to reference the
     * added entry.  Once the entry is added, it is owned by the Event object.
     */
    uint64_t AddEntry(google::protobuf::Message *entry, std::string tag = "");
    /** GetEntry takes an entry ID and returns the corresponding entry.  The
     * returned entries are still owned by the Event object.
     */
    google::protobuf::Message *GetEntry(uint64_t id);
    /** TagEntry adds a tag to an entry that has already been added, identified
     * by its ID.
     */
    void TagEntry(uint64_t id, std::string tag);
    /** UntagEntry removes a tag from a specified entry.
     */
    void UntagEntry(uint64_t id, std::string tag);
    /** RemoveEntry removes an entry from the event.
     */
    void RemoveEntry(uint64_t id);
    /** Tags returns a list of tags that exist in the event.
     */
    std::vector<std::string> Tags();
    /** TaggedEntries tages a tag string and returns a list of entry IDs that
     * the tag references.
     */
    std::vector<uint64_t> TaggedEntries(std::string tag);
    /** AllEntries returns IDs for all entries in the event.
     */
    std::vector<uint64_t> AllEntries();
    /** EntryTags performs a reverse lookup of tags that point to an entry.
     * This is a relatively expensive search.
     */
    std::vector<std::string> EntryTags(uint64_t id);
    /** DeleteTag removes a tag from the Event.
     */
    void DeleteTag(std::string tag);
    /** Free returns an allocated and cleared entry Message of the type
     * described by the given Descriptor, or NULL if no already-allocated
     * messages of this type are available.
     */
    google::protobuf::Message *Free(const google::protobuf::Descriptor *desc);
    /** Metadata returns a mapping from a string key to a pointer to a string
     * that contains the metadata, by reference.  These metadata are all the
     * entries received on the stream up to this Event.
     */
    const std::map<std::string, std::shared_ptr<std::string>> &Metadata() { return metadata; }

    /** String returns a human-readable string representing the event.
     */
    std::string String();
    /** SerializeToString stores the event in the given string in protobuf wire
     * format.
     */
    bool SerializeToString(std::string *output) {
        FlushCache();
        return getProto()->SerializeToString(output);
    }
    /** FlushCache forces all event entries to be serialized.  This is useful
     * for putting the main serialization load into parallel threads before
     * aggregating the events into an output stream.
     */
    void FlushCache();
    /** Clear prepares the Event for data from a new event.
     */
    void Clear();

    Event &operator=(const Event &event);

   private:
    uint64_t getTypeID(google::protobuf::Message *entry);
    const google::protobuf::Descriptor *getDescriptor(uint64_t typeID);
    void tagCleanup();

    friend class Writer;
    friend class Reader;

    proto::Event *getProto();

    proto::Event *eventProto;
    std::map<std::string, uint64_t> revTypeLookup;
    std::map<uint64_t, google::protobuf::Message *> entryCache;
    std::map<uint64_t, const google::protobuf::Descriptor *> descriptorCache;
    std::map<std::string, std::shared_ptr<std::string>> metadata;
    bool dirtyTags;

    std::map<const google::protobuf::Descriptor *, std::vector<google::protobuf::Message *>> store;
};

const class UnknownMessageTypeError : public std::exception {
    virtual const char *what() const throw() { return "Unknown message type"; }
} unknownMessageTypeError;
}  // namespace proio

#endif  // PROIO_EVENT_H
