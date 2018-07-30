#ifndef PROIO_READER_H
#define PROIO_READER_H

#include <cstring>
#include <mutex>
#include <string>

#include <google/protobuf/io/zero_copy_stream_impl.h>
#include <lz4frame.h>

#include "event.h"

namespace proio {
class BucketInputStream : public google::protobuf::io::ZeroCopyInputStream {
   public:
    BucketInputStream(uint64_t size);
    virtual ~BucketInputStream();

    bool Next(const void **data, int *size);
    void BackUp(int count);
    bool Skip(int count);
    google::protobuf::int64 ByteCount() const;

    uint8_t *Bytes();
    uint64_t BytesRemaining();
    void Reset(uint64_t size);
    uint64_t Reset(google::protobuf::io::ZeroCopyInputStream &stream);
    uint64_t Reset(LZ4F_dctx *dctxPtr, BucketInputStream *compBucket);

   private:
    uint64_t offset;
    std::vector<uint8_t> bytes;
    uint64_t size;
};

/** Reader for proio files
 */
class Reader : public std::mutex {
   public:
    /** Constructor for providing a file descriptor
     */
    Reader(int fd);
    /** Constructor that creates a file descriptor from the path of an existing
     * file.
     */
    Reader(std::string filename);
    virtual ~Reader();

    /** Next returns the next available Event.  This function takes two
     * important optional arguments.  The first argument is a pointer to a
     * recycled Event to be cleared and filled by the Reader.  The second
     * argument is an option for filling only metadata; this is useful for
     * scanning a file on disk, since only bucket headers need to be read for
     * this.
     */
    Event *Next(Event *recycledEvent = NULL, bool metadataOnly = false);
    /** Next copies the next event data in protobuf wire format into the given
     * string.
     */
    bool Next(std::string *data);
    /** Skip skips the next nEvents events.
     */
    uint64_t Skip(uint64_t nEvents);
    /** SeekToStart sends the reader to the beginning of the stream if it is
     * seekable.
     */
    void SeekToStart();

   private:
    void initBucket();
    void readFromBucket(Event *event);
    void readFromBucket(std::string *data);
    void readHeader();
    void readBucket();
    uint64_t syncToMagic(google::protobuf::io::CodedInputStream *stream);

    BucketInputStream *compBucket;
    google::protobuf::io::FileInputStream *fileStream;
    int fd;
    bool closeFDOnDelete;
    proto::BucketHeader *bucketHeader;
    uint64_t bucketEventsRead;
    uint64_t bucketIndex;
    LZ4F_dctx *dctxPtr;
    BucketInputStream *bucket;
    std::map<std::string, std::shared_ptr<std::string>> metadata;
};

const class FileOpenError : public std::exception {
    virtual const char *what() const throw() { return "Failed to open file for reading"; }
} fileOpenError;

const class DeserializationError : public std::exception {
    virtual const char *what() const throw() { return "Failed to deserialize message"; }
} deserializationError;

const class CorruptBucketError : public std::exception {
    virtual const char *what() const throw() { return "Bucket is corrupt"; }
} corruptBucketError;

const class BadLZ4FrameError : public std::exception {
    virtual const char *what() const throw() { return "Bad LZ4 frame"; }
} badLZ4FrameError;

const class SeekError : public std::exception {
    virtual const char *what() const throw() { return "Failed to seek file"; }
} seekError;

const class IOError : public std::exception {
    virtual const char *what() const throw() { return "Unexpected IO Error"; }
} ioError;
}  // namespace proio

#endif  // PROIO_READER_H
