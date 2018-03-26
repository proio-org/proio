#include <fcntl.h>

#include <google/protobuf/io/gzip_stream.h>
#include <lz4frame.h>
//#include <lz4hc.h>

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

    pthread_cond_wait(&streamWriteReadyCond, &streamWriteReadyMutex);
    streamWriteJob = NULL;
    pthread_mutex_lock(&streamWriteJobMutex);
    pthread_cond_signal(&streamWriteJobCond);
    pthread_mutex_unlock(&streamWriteJobMutex);
    pthread_mutex_unlock(&streamWriteReadyMutex);

    pthread_join(streamWriteThread, NULL);
    pthread_cond_destroy(&streamWriteReadyCond);
    pthread_mutex_destroy(&streamWriteReadyMutex);
    pthread_cond_destroy(&streamWriteJobCond);
    pthread_mutex_destroy(&streamWriteJobMutex);

    delete bucket;
    delete fileStream;
    delete compBucket;
}

void Writer::Flush() {
    if (bucketEvents == 0) return;

    pthread_cond_wait(&streamWriteReadyCond, &streamWriteReadyMutex);
    compBucket->Reset();
    switch (compression) {
        case LZ4: {
            LZ4F_frameInfo_t info;
            std::memset(&info, 0, sizeof(info));
            LZ4F_preferences_t prefs;
            std::memset(&prefs, 0, sizeof(prefs));
            prefs.frameInfo = info;
            prefs.compressionLevel = 0;  // LZ4HC_CLEVEL_MAX;
            size_t compBound = LZ4F_compressFrameBound(bucket->ByteCount(), &prefs);
            compBucket->Reset(compBound);
            size_t nWritten = LZ4F_compressFrame(compBucket->Bytes(), compBound, bucket->Bytes(),
                                                 bucket->ByteCount(), &prefs);
            if (LZ4F_isError(nWritten)) throw lz4FrameCreationError;
            compBucket->SetOffset(nWritten);
        } break;
        case GZIP: {
            io::GzipOutputStream *gzipStream = new io::GzipOutputStream(compBucket);
            bucket->WriteTo(gzipStream);
            delete gzipStream;
        } break;
        default:
            BucketOutputStream *tmpBucket = bucket;
            bucket = compBucket;
            compBucket = tmpBucket;
    }

    auto header = new proto::BucketHeader();
    header->set_nevents(bucketEvents);
    header->set_bucketsize(compBucket->ByteCount());
    header->set_compression(compression);

    streamWriteJob = new WriteJob;
    streamWriteJob->compBucket = compBucket;
    streamWriteJob->header = header;
    streamWriteJob->fileStream = fileStream;
    pthread_mutex_lock(&streamWriteJobMutex);
    pthread_cond_signal(&streamWriteJobCond);
    pthread_mutex_unlock(&streamWriteJobMutex);

    bucket->Reset();
    bucketEvents = 0;
}

void Writer::Push(Event *event) {
    event->flushCache();
    proto::Event *proto = event->getProto();

    auto stream = new io::CodedOutputStream(bucket);
#if GOOGLE_PROTOBUF_VERSION >= 3004000
    stream->WriteLittleEndian32((uint32_t)proto->ByteSizeLong());
#else
    stream->WriteLittleEndian32((uint32_t)proto->ByteSize());
#endif
    if (!proto->SerializeToCodedStream(stream)) throw serializationError;
    delete stream;

    bucketEvents++;

    if (bucket->ByteCount() > bucketDumpThres) Flush();
}

void Writer::SetCompression(Compression comp) { compression = comp; }

void Writer::initBucket() {
    bucket = new BucketOutputStream();
    bucketEvents = 0;

    compression = LZ4;
    compBucket = new BucketOutputStream();

    pthread_mutex_init(&streamWriteJobMutex, NULL);
    pthread_cond_init(&streamWriteJobCond, NULL);
    pthread_mutex_init(&streamWriteReadyMutex, NULL);
    pthread_cond_init(&streamWriteReadyCond, NULL);

    pthread_mutex_lock(&streamWriteReadyMutex);
    pthread_create(&streamWriteThread, NULL, Writer::streamWrite, this);
}

void *Writer::streamWrite(void *writerVoid) {
    auto writer = (Writer *)writerVoid;

    pthread_mutex_lock(&writer->streamWriteJobMutex);
    while (true) {
        pthread_mutex_lock(&writer->streamWriteReadyMutex);
        pthread_cond_signal(&writer->streamWriteReadyCond);
        pthread_mutex_unlock(&writer->streamWriteReadyMutex);
        pthread_cond_wait(&writer->streamWriteJobCond, &writer->streamWriteJobMutex);

        auto job = writer->streamWriteJob;
        if (job) {
            auto stream = new io::CodedOutputStream(job->fileStream);
            stream->WriteRaw(magicBytes, 16);
#if GOOGLE_PROTOBUF_VERSION >= 3004000
            stream->WriteLittleEndian32((uint32_t)job->header->ByteSizeLong());
#else
            stream->WriteLittleEndian32((uint32_t)job->header->ByteSize());
#endif
            if (!job->header->SerializeToCodedStream(stream)) throw serializationError;
            stream->WriteRaw(job->compBucket->Bytes(), job->compBucket->ByteCount());
            delete stream;

            delete job->header;
            delete job;
            writer->streamWriteJob = NULL;
        } else
            break;
    }
    pthread_mutex_unlock(&writer->streamWriteJobMutex);

    return NULL;
}

BucketOutputStream::BucketOutputStream() { offset = 0; }

BucketOutputStream::~BucketOutputStream() { ; }

inline bool BucketOutputStream::Next(void **data, int *size) {
    if (bytes.size() - offset < minBucketWriteWindow) bytes.resize(offset + minBucketWriteWindow);
    *data = &bytes[offset];
    *size = bytes.size() - offset;
    offset = bytes.size();
    return true;
}

inline void BucketOutputStream::BackUp(int count) { offset -= count; }

inline int64 BucketOutputStream::ByteCount() const { return offset; }

inline bool BucketOutputStream::AllowsAliasing() { return false; }

uint8_t *BucketOutputStream::Bytes() { return &bytes[0]; }

void BucketOutputStream::Reset() { offset = 0; }

void BucketOutputStream::Reset(uint64_t size) {
    offset = 0;
    if (bytes.size() < size) bytes.resize(size);
}

void BucketOutputStream::WriteTo(io::ZeroCopyOutputStream *stream) {
    uint8_t *data;
    int size;
    uint64_t bytesWritten = 0;
    while (stream->Next((void **)&data, &size)) {
        uint64_t bytesLeft = offset - bytesWritten;
        uint64_t bytesToCopy = (bytesLeft < size) ? bytesLeft : size;
        std::memcpy(data, Bytes() + bytesWritten, bytesToCopy);
        bytesLeft -= bytesToCopy;
        bytesWritten += bytesToCopy;
        if (bytesToCopy < size) stream->BackUp(size - bytesToCopy);
        if (bytesLeft == 0) break;
    }
}

void BucketOutputStream::SetOffset(uint64_t offset) { this->offset = offset; }
