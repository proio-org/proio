#ifndef PROIO_WRITER_H
#define PROIO_WRITER_H

#include <condition_variable>
#include <cstring>
#include <mutex>
#include <string>
#include <thread>

#include <google/protobuf/io/zero_copy_stream_impl.h>

#include "event.h"
#include "proio.pb.h"

namespace proio {
typedef proto::BucketHeader_CompType Compression;
const Compression LZ4 = proto::BucketHeader::LZ4;
const Compression GZIP = proto::BucketHeader::GZIP;
const Compression UNCOMPRESSED = proto::BucketHeader::NONE;

const std::size_t minBucketWriteWindow = 0x100000;

class BucketOutputStream : public google::protobuf::io::ZeroCopyOutputStream {
   public:
    BucketOutputStream();
    virtual ~BucketOutputStream();

    bool Next(void **data, int *size);
    void BackUp(int count);
    google::protobuf::int64 ByteCount() const;
    bool AllowsAliasing();

    uint8_t *Bytes();
    void Reset();
    void Reset(uint64_t size);
    void WriteTo(google::protobuf::io::ZeroCopyOutputStream *stream);
    void SetOffset(uint64_t offset);

   private:
    std::vector<uint8_t> bytes;
    uint64_t offset;
};

/** Writer for proio files
 */
class Writer : public std::mutex {
   public:
    /** Constructor for providing a file descriptor
     */
    Writer(int fd);
    /** Constructor that creates a file descriptor from a file path,
     * overwriting any existing file
     */
    Writer(std::string filename);
    virtual ~Writer();

    /** Flush compresses and flushes all buffered (bucket) data to the output
     * stream.  This is automatically called upon destruction of the Writer
     * object, and when the bucket dump threshold is reached.  This function is
     * asynchronous, meaning that it will return before the flush to the output
     * stream is actually complete.  Synchronization is forced on destruction.
     */
    void Flush();
    /** Push takes an Event and serializes it into the output bucket.
     */
    void Push(Event *event);
    /** PushMetadata takes a string key and string data set and pushes it into
     * the stream.  If Events exist in the current bucket, the bucket is
     * flushed first.
     */
    void PushMetadata(std::string name, std::string &data);
    /** PushMetadata takes a string key and null-terminated const char array by
     * pointer and pushes it into the stream.  If Events exist in the current
     * bucket, the bucket is flushed first.
     */
    void PushMetadata(std::string name, const char *data);
    /** SetCompression sets the compression type to use for future output
     * buckets.  One of: LZ4, GZIP, or UNCOMPRESSED.
     */
    void SetCompression(Compression comp = GZIP) { compression = comp; }
    /** SetBucketDumpThreshold sets the threshold uncompressed bucket size for
     * automatic compression and output (dump).  I.e., once the size of the
     * uncompressed bucket in memory reaches this threshold, Flush() will be
     * called.  Flush() can also be manually called at any time.
     */
    void SetBucketDumpThreshold(uint64_t thres = 0x1000000) { bucketDumpThres = thres; }

   private:
    BucketOutputStream *bucket;
    google::protobuf::io::FileOutputStream *fileStream;
    uint64_t bucketEvents;
    Compression compression;
    BucketOutputStream *compBucket;
    uint64_t bucketDumpThres;
    proto::BucketHeader *header;
    std::map<std::string, std::shared_ptr<std::string>> metadata;

    void initBucket();

    std::thread streamWriteThread;
    typedef struct {
        bool isValid;

        BucketOutputStream *compBucket;
        proto::BucketHeader *header;
        google::protobuf::io::FileOutputStream *fileStream;

        std::mutex doJobMutex;
        std::condition_variable doJobCond;
        std::mutex workerReadyMutex;
        std::condition_variable workerReadyCond;
    } WriteJob;
    WriteJob streamWriteJob;
    std::unique_lock<std::mutex> workerReadyLock;

    static void streamWrite(WriteJob *job);
};

const uint8_t magicBytes[] = {0xe1, 0xc1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
                              0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00};

const class SerializationError : public std::exception {
    virtual const char *what() const throw() { return "Failed to serialize message"; }
} serializationError;

const class FileCreationError : public std::exception {
    virtual const char *what() const throw() { return "Failed to creating file for writing"; }
} fileCreationError;

const class LZ4FrameCreationError : public std::exception {
    virtual const char *what() const throw() { return "Failed to create LZ4 frame"; }
} lz4FrameCreationError;
}  // namespace proio

#endif  // PROIO_WRITER_H
