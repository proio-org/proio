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
    BucketInputStream(uint64_t size) {
        offset = 0;
        bytes.resize(size);
        this->size = size;
    }
    virtual ~BucketInputStream() { ; }

    bool Next(const void **data, int *size) {
        *data = &bytes[offset];
        *size = this->size - offset;
        offset = this->size;
        if (*size == 0) return false;
        return true;
    }
    void BackUp(int count) {
        offset -= count;
        if (offset < 0) offset = 0;
    }
    bool Skip(int count) {
        offset += count;
        if (offset > size) {
            offset = size;
            return false;
        }
        return true;
    }
    google::protobuf::int64 ByteCount() const { return offset; }

    uint8_t *Bytes() { return &bytes[0]; }
    uint64_t BytesRemaining() { return size - offset; }
    void Reset(uint64_t size) {
        offset = 0;
        if (bytes.size() < size) bytes.resize(size);
        this->size = size;
    }
    uint64_t Reset(google::protobuf::io::ZeroCopyInputStream &stream) {
        Reset(0);
        uint8_t *data;
        int size;
        while (stream.Next((const void **)&data, &size)) {
            offset = this->size;
            this->size += size;
            if (this->size > bytes.size()) bytes.resize(size);
            std::memcpy(&bytes[offset], data, size);
        }
        offset = 0;
        return this->size;
    }

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
