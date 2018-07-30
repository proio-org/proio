#include <fcntl.h>
#include <sys/types.h>

#include <google/protobuf/io/gzip_stream.h>
#include <lz4.h>

#include "reader.h"
#include "writer.h"

using namespace proio;
using namespace google::protobuf;

Reader::Reader(int fd) {
    this->fd = fd;
    fileStream = new io::FileInputStream(fd);
    closeFDOnDelete = false;

    initBucket();
}

Reader::Reader(std::string filename) {
    fd = open(filename.c_str(), O_RDONLY);
    if (fd == -1) throw fileOpenError;
    fileStream = new io::FileInputStream(fd);
    closeFDOnDelete = true;

    initBucket();
}

Reader::~Reader() {
    if (bucketHeader) delete bucketHeader;
    delete compBucket;
    LZ4F_freeDecompressionContext(dctxPtr);
    delete bucket;
    delete fileStream;
    if (closeFDOnDelete) close(fd);
}

Event *Reader::Next(Event *event, bool metaOnly) {
    if (event)
        event->Clear();
    else
        event = new Event();

    while (!bucketHeader || bucketIndex >= bucketHeader->nevents()) {
        if (bucketHeader) bucketIndex -= bucketHeader->nevents();
        readHeader();
        if (!bucketHeader) return NULL;
    }
    event->metadata = metadata;
    if (!metaOnly) {
        if (bucket->BytesRemaining() == 0) readBucket();
        readFromBucket(event);
    } else
        bucketIndex++;

    return event;
}

bool Reader::Next(std::string *data) {
    if (!data) return false;

    while (!bucketHeader || bucketIndex >= bucketHeader->nevents()) {
        if (bucketHeader) bucketIndex -= bucketHeader->nevents();
        readHeader();
        if (!bucketHeader) return false;
    }

    if (bucket->BytesRemaining() == 0) readBucket();
    readFromBucket(data);

    return true;
}

uint64_t Reader::Skip(uint64_t nEvents) {
    uint64_t nSkipped = 0;

    uint64_t startIndex = bucketIndex;
    bucketIndex += nEvents;
    while (!bucketHeader || bucketIndex >= bucketHeader->nevents()) {
        if (bucketHeader) {
            uint64_t nBucketEvents = bucketHeader->nevents();
            bucketIndex -= nBucketEvents;
            nSkipped += nBucketEvents - startIndex;
            if (nBucketEvents > 0 && bucket->BytesRemaining() == 0)
                if (!fileStream->Skip(bucketHeader->bucketsize())) throw ioError;
        }
        readHeader();
        if (!bucketHeader) return nSkipped;
        startIndex = 0;
    }
    nSkipped += bucketIndex - startIndex;

    return nSkipped;
}

void Reader::SeekToStart() {
    delete fileStream;
    if (lseek(fd, 0, SEEK_SET) == -1) throw seekError;
    fileStream = new io::FileInputStream(fd);
    metadata.clear();
    bucketIndex = 0;
    readHeader();
}

void Reader::initBucket() {
    compBucket = new BucketInputStream(0);
    bucketHeader = NULL;
    bucketEventsRead = 0;
    bucketIndex = 0;
    LZ4F_createDecompressionContext(&dctxPtr, LZ4F_VERSION);
    bucket = new BucketInputStream(0);
}

void Reader::readFromBucket(Event *event) {
    auto stream = new io::CodedInputStream(bucket);

    while (bucketEventsRead <= bucketIndex) {
        uint32_t protoSize;
        if (!stream->ReadLittleEndian32(&protoSize)) {
            delete stream;
            throw corruptBucketError;
        }

        if (event && bucketEventsRead == bucketIndex) {
            auto eventLimit = stream->PushLimit(protoSize);
            auto eventProto = event->getProto();
            if (!eventProto->MergeFromCodedStream(stream) || !stream->ConsumedEntireMessage()) {
                delete stream;
                throw deserializationError;
            }
            stream->PopLimit(eventLimit);
        } else if (!stream->Skip(protoSize)) {
            delete stream;
            throw corruptBucketError;
        }

        bucketEventsRead++;
    }
    bucketIndex++;

    delete stream;
}

void Reader::readFromBucket(std::string *data) {
    auto stream = new io::CodedInputStream(bucket);

    while (bucketEventsRead <= bucketIndex) {
        uint32_t protoSize;
        if (!stream->ReadLittleEndian32(&protoSize)) {
            delete stream;
            throw corruptBucketError;
        }

        if (data && bucketEventsRead == bucketIndex) {
            data->resize(protoSize);
            if (!stream->ReadString(data, protoSize)) {
                delete stream;
                throw corruptBucketError;
            }
        } else if (!stream->Skip(protoSize)) {
            delete stream;
            throw corruptBucketError;
        }

        bucketEventsRead++;
    }
    bucketIndex++;

    delete stream;
}

void Reader::readHeader() {
    if (bucketHeader) {
        delete bucketHeader;
        bucketHeader = NULL;
    }
    bucketEventsRead = 0;
    compBucket->Reset(0);
    bucket->Reset(0);

    auto stream = new io::CodedInputStream(fileStream);
    syncToMagic(stream);
    uint32_t headerSize;
    if (!stream->ReadLittleEndian32(&headerSize)) {
        delete stream;
        return;
    }

    auto headerLimit = stream->PushLimit(headerSize);
    bucketHeader = new proto::BucketHeader;
    if (!bucketHeader->MergeFromCodedStream(stream) || !stream->ConsumedEntireMessage()) {
        delete stream;
        throw deserializationError;
    }
    stream->PopLimit(headerLimit);

    // Set metadata for future events
    for (auto keyValuePair : bucketHeader->metadata())
        metadata[keyValuePair.first] = std::make_shared<std::string>(keyValuePair.second);

    delete stream;
}

void Reader::readBucket() {
    auto stream = new io::CodedInputStream(fileStream);

    uint64_t bucketSize = bucketHeader->bucketsize();
    compBucket->Reset(bucketSize);
    if (!stream->ReadRaw(compBucket->Bytes(), bucketSize)) {
        delete stream;
        throw corruptBucketError;
    }

    delete stream;

    switch (bucketHeader->compression()) {
        case LZ4: {
            bucket->Reset(dctxPtr, compBucket);
            break;
        }
        case GZIP: {
            io::GzipInputStream *gzipStream = new io::GzipInputStream(compBucket);
            bucket->Reset(*gzipStream);
            delete gzipStream;
            break;
        }
        default:
            BucketInputStream *tmpBucket = bucket;
            bucket = compBucket;
            compBucket = tmpBucket;
    }
}

uint64_t Reader::syncToMagic(io::CodedInputStream *stream) {
    uint8_t num;
    uint64_t nRead = 0;

    while (stream->ReadRaw(&num, 1)) {
        nRead++;

        if (num == magicBytes[0]) {
            bool goodSeq = true;

            for (int i = 1; i < 16; i++) {
                if (!stream->ReadRaw(&num, 1)) break;
                nRead++;

                if (num != magicBytes[i]) {
                    goodSeq = false;
                    break;
                }
            }
            if (goodSeq) break;
        }
    }
    return nRead;
}

BucketInputStream::BucketInputStream(uint64_t size) {
    offset = 0;
    bytes.resize(size);
    this->size = size;
}

BucketInputStream::~BucketInputStream() { ; }

inline bool BucketInputStream::Next(const void **data, int *size) {
    *data = &bytes[offset];
    *size = this->size - offset;
    offset = this->size;
    if (*size == 0) return false;
    return true;
}

inline void BucketInputStream::BackUp(int count) { offset -= count; }

inline bool BucketInputStream::Skip(int count) {
    offset += count;
    if (offset > size) {
        offset = size;
        return false;
    }
    return true;
}

inline int64 BucketInputStream::ByteCount() const { return offset; }

uint8_t *BucketInputStream::Bytes() { return &bytes[0]; }

uint64_t BucketInputStream::BytesRemaining() { return size - offset; }

void BucketInputStream::Reset(uint64_t size) {
    offset = 0;
    if (bytes.size() < size) bytes.resize(size);
    this->size = size;
}

uint64_t BucketInputStream::Reset(io::ZeroCopyInputStream &stream) {
    Reset(0);
    uint8_t *data;
    int size;
    while (stream.Next((const void **)&data, &size)) {
        offset = this->size;
        this->size += size;
        if (this->size > bytes.size()) bytes.resize(this->size);
        std::memcpy(&bytes[offset], data, size);
    }
    offset = 0;
    return this->size;
}

uint64_t BucketInputStream::Reset(LZ4F_dctx *dctxPtr, BucketInputStream *compBucket) {
    offset = 0;
    size = bytes.size();
    if (size == 0) Reset(minBucketWriteWindow);
    int srcSize;
    uint8_t *srcBuffer;
    compBucket->Next((const void **)&srcBuffer, &srcSize);
    int srcBytesRemaining = srcSize;
    int dstSize;
    uint8_t *dstBuffer;
    size_t hint;
    while (srcBytesRemaining > 0) {
        Next((const void **)&dstBuffer, &dstSize);
        size_t srcSizeTmp = srcSize;
        size_t dstSizeTmp = dstSize;
        hint = LZ4F_decompress(dctxPtr, dstBuffer, &dstSizeTmp, srcBuffer, &srcSizeTmp, NULL);
        if (LZ4F_isError(hint)) throw badLZ4FrameError;
        srcBytesRemaining -= srcSizeTmp;
        BackUp(dstSize - dstSizeTmp);
        if (offset == size) {
            size += minBucketWriteWindow;
            bytes.resize(size);
        }
        compBucket->BackUp(srcBytesRemaining);
        compBucket->Next((const void **)&srcBuffer, &srcSize);
    }
    size = offset;
    offset = 0;
    if (hint != 0) {
        LZ4F_resetDecompressionContext(dctxPtr);
        throw badLZ4FrameError;
    }
    return size;
}
