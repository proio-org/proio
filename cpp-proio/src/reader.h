#ifndef PROIO_READER_H
#define PROIO_READER_H

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

   private:
    uint64_t offset;
    std::vector<uint8_t> bytes;
    uint64_t size;
};

class Reader {
   public:
    Reader(int fd);
    Reader(std::string filename);
    virtual ~Reader();

    Event *Next();

   private:
    void initBucket();
    Event *readFromBucket(bool doMerge = true);
    uint64_t readBucket(uint64_t maxSkipEvents = 0);
    uint64_t syncToMagic(google::protobuf::io::CodedInputStream &);

    BucketInputStream *compBucket;
    google::protobuf::io::FileInputStream *fileStream;

    uint64_t bucketEventsRead;
    proto::BucketHeader *bucketHeader;

    LZ4F_dctx *dctxPtr;
    BucketInputStream *bucket;
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
}  // namespace proio

#endif  // PROIO_READER_H
