#include <fcntl.h>

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
}

void Writer::Flush() {
    if (bucketEvents == 0) return;
    auto stream = new io::CodedOutputStream(fileStream);

    uint8 *bytes = bucket->Bytes();
    uint64 nBytes = bucket->ByteCount();

    auto header = new proto::BucketHeader();
    header->set_nevents(bucketEvents);
    header->set_bucketsize(nBytes);
    header->set_compression(UNCOMPRESSED);

    stream->WriteRaw(magicBytes, 16);
#if GOOGLE_PROTOBUF_VERSION >= 3004000
    stream->WriteLittleEndian32((uint32)header->ByteSizeLong());
#else
    stream->WriteLittleEndian32((uint32)header->ByteSize());
#endif
    if (!header->SerializeToCodedStream(stream)) {
        delete stream;
        throw serializationError;
    }
    stream->WriteRaw(bytes, nBytes);

    bucket->Reset();
    bucketEvents = 0;

    delete stream;
}

void Writer::Push(Event *event) {
    auto stream = new io::CodedOutputStream(bucket);

    event->flushCollCache();
    proto::Event *proto = event->getProto();
#if GOOGLE_PROTOBUF_VERSION >= 3004000
    stream->WriteLittleEndian32((uint32)proto->ByteSizeLong());
#else
    stream->WriteLittleEndian32((uint32)proto->ByteSize());
#endif
    if (!proto->SerializeToCodedStream(stream)) {
        delete stream;
        throw serializationError;
    }

    bucketEvents++;
    delete stream;

    if (bucket->ByteCount() > bucketDumpSize) {
        Flush();
    }
}

void Writer::SetCompression(Compression comp) { compression = comp; }

void Writer::initBucket() {
    bucket = new BucketOutputStream();
    bucketEvents = 0;

    compression = UNCOMPRESSED;
}
