#include <fcntl.h>

#include <google/protobuf/io/gzip_stream.h>
#include <lz4frame.h>

#include "writer.h"

using namespace proio;
using namespace google::protobuf;

Writer::Writer(int fd) {
    fileStream = new io::FileOutputStream(fd);
    fileStream->SetCloseOnDelete(false);

    initBucket();
}

Writer::Writer(std::string filename) {
    int fd = open(filename.c_str(), O_WRONLY | O_CREAT | O_TRUNC, S_IRUSR | S_IWUSR | S_IRGRP | S_IROTH);
    if (fd == -1) throw fileCreationError;
    fileStream = new io::FileOutputStream(fd);
    fileStream->SetCloseOnDelete(true);

    initBucket();
}

Writer::~Writer() {
    Flush();

    delete bucket;
    delete fileStream;
    delete compBucket;
}

void Writer::Flush() {
    if (bucketEvents == 0) return;
    io::CodedOutputStream stream(fileStream);

    compBucket->Reset();
    switch (compression) {
        case LZ4: {
            LZ4F_frameInfo_t info;
            info.contentSize = bucket->ByteCount();
            LZ4F_preferences_t prefs;
            prefs.frameInfo = info;
            size_t compBound = LZ4F_compressFrameBound(bucket->ByteCount(), &prefs);
            compBucket->Reset(compBound);
            compBucket->SetOffset(LZ4F_compressFrame(compBucket->Bytes(), compBound, bucket->Bytes(),
                                                     bucket->ByteCount(), &prefs));
            break;
        }
        case GZIP: {
            io::GzipOutputStream *gzipStream = new io::GzipOutputStream(compBucket);
            bucket->WriteTo(gzipStream);
            delete gzipStream;
            break;
        }
        default:
            BucketOutputStream *tmpBucket = bucket;
            bucket = compBucket;
            compBucket = tmpBucket;
    }

    auto header = new proto::BucketHeader();
    header->set_nevents(bucketEvents);
    header->set_bucketsize(compBucket->ByteCount());
    header->set_compression(compression);

    stream.WriteRaw(magicBytes, 16);
#if GOOGLE_PROTOBUF_VERSION >= 3004000
    stream.WriteLittleEndian32((uint32_t)header->ByteSizeLong());
#else
    stream.WriteLittleEndian32((uint32_t)header->ByteSize());
#endif
    if (!header->SerializeToCodedStream(&stream)) throw serializationError;
    stream.WriteRaw(compBucket->Bytes(), compBucket->ByteCount());

    bucket->Reset();
    bucketEvents = 0;
}

void Writer::Push(Event *event) {
    io::CodedOutputStream stream(bucket);

    event->flushCollCache();
    proto::Event *proto = event->getProto();
#if GOOGLE_PROTOBUF_VERSION >= 3004000
    stream.WriteLittleEndian32((uint32_t)proto->ByteSizeLong());
#else
    stream.WriteLittleEndian32((uint32_t)proto->ByteSize());
#endif
    if (!proto->SerializeToCodedStream(&stream)) throw serializationError;

    bucketEvents++;

    if (bucket->ByteCount() > bucketDumpSize) {
        Flush();
    }
}

void Writer::SetCompression(Compression comp) { compression = comp; }

void Writer::initBucket() {
    bucket = new BucketOutputStream();
    bucketEvents = 0;

    compression = LZ4;
    compBucket = new BucketOutputStream();
}
