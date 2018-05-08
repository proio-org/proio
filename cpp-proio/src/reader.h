#ifndef PROIO_READER_H
#define PROIO_READER_H

#include <pthread.h>
#include <cstring>
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
class Reader {
   public:
    /** Constructor for providing a file descriptor
     */
    Reader(int fd);
    /** Constructor that creates a file descriptor from the path of an existing
     * file.
     */
    Reader(std::string filename);
    virtual ~Reader();

    /** Next returns the next available Event.
     */
    Event *Next();
    /** NextHeader returns the next bucket header.  This is useful for scanning
     * a stream.  The corresponding bucket is discarded.
     */
    proto::BucketHeader *NextHeader();
    /** Skip skips the next nEvents events.
     */
    uint64_t Skip(uint64_t nEvents);
    /** SeekToStart sends the reader to the beginning of the stream if it is
     * seekable.
     */
    void SeekToStart();

   private:
    void initBucket();
    Event *readFromBucket(bool doMerge = true);
    uint64_t readBucket(uint64_t maxSkipEvents = 0);
    uint64_t syncToMagic(google::protobuf::io::CodedInputStream *stream);

    BucketInputStream *compBucket;
    google::protobuf::io::FileInputStream *fileStream;
    int fd;
    bool closeFDOnDelete;
    uint64_t bucketEventsRead;
    proto::BucketHeader *bucketHeader;
    LZ4F_dctx *dctxPtr;
    BucketInputStream *bucket;
    std::map<std::string, std::shared_ptr<std::string>> metadata;

    pthread_mutex_t mutex;
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
}  // namespace proio

#endif  // PROIO_READER_H
