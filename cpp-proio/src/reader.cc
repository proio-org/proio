#include <fcntl.h>

#include "reader.h"
#include "writer.h"

using namespace proio;
using namespace google::protobuf;

Reader::Reader(int fd) {
    fileStream = new io::FileInputStream(fd);
    fileStream->SetCloseOnDelete(false);

    initBucket();
}

Reader::Reader(std::string filename) {
    int fd = open(filename.c_str(), O_RDONLY);
    if (fd == -1) throw fileOpenError;
    fileStream = new io::FileInputStream(fd);
    fileStream->SetCloseOnDelete(true);

    initBucket();
}

Reader::~Reader() {
    if (bucketHeader) delete bucketHeader;
    delete bucket;
    delete fileStream;
}

Event *Reader::Next() { return readFromBucket(); }

void Reader::initBucket() {
    bucket = new BucketInputStream(0);
    bucketEventsRead = 0;
    bucketHeader = NULL;
}

Event *Reader::readFromBucket(bool doMerge) {
    if (bucket->BytesRemaining() == 0) readBucket();
    io::CodedInputStream stream(bucket);

    uint32_t protoSize;
    if (!stream.ReadLittleEndian32(&protoSize)) return NULL;

    bucketEventsRead++;
    if (doMerge) {
        auto eventLimit = stream.PushLimit(protoSize);
        auto eventProto = new proto::Event;
        if (!eventProto->MergeFromCodedStream(&stream) || !stream.ConsumedEntireMessage())
            throw deserializationError;
        return new Event(eventProto);
    } else {
        if (!stream.Skip(protoSize)) throw corruptBucketError;
        return NULL;
    }
}

uint64_t Reader::readBucket(uint64_t maxSkipEvents) {
    io::CodedInputStream stream(fileStream);
    syncToMagic(stream);

    bucketEventsRead = 0;
    bucket->Reset(0);

    uint32_t headerSize;
    if (!stream.ReadLittleEndian32(&headerSize)) return 0;

    auto headerLimit = stream.PushLimit(headerSize);
    if (bucketHeader) delete bucketHeader;
    bucketHeader = new proto::BucketHeader;
    if (!bucketHeader->MergeFromCodedStream(&stream) || !stream.ConsumedEntireMessage())
        throw deserializationError;
    stream.PopLimit(headerLimit);

    uint64_t bucketSize = bucketHeader->bucketsize();
    if (bucketHeader->nevents() > maxSkipEvents) {
        bucket->Reset(bucketSize);
        if (!stream.ReadRaw(bucket->Bytes(), bucketSize)) throw corruptBucketError;
    } else {
        if (!stream.Skip(bucketSize)) throw corruptBucketError;
        return bucketHeader->nevents();
    }
}

uint64_t Reader::syncToMagic(io::CodedInputStream &stream) {
    uint8_t num;
    uint64_t nRead = 0;

    while (stream.ReadRaw(&num, 1)) {
        nRead++;

        if (num == magicBytes[0]) {
            bool goodSeq = true;

            for (int i = 1; i < 16; i++) {
                if (!stream.ReadRaw(&num, 1)) break;
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
