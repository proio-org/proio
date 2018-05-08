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
    pthread_mutex_lock(&mutex);
    pthread_mutex_unlock(&mutex);
    pthread_mutex_destroy(&mutex);

    Flush();

    pthread_cond_wait(&streamWriteJob.workerReadyCond, &streamWriteJob.workerReadyMutex);
    streamWriteJob.isValid = false;
    pthread_mutex_lock(&streamWriteJob.doJobMutex);
    pthread_cond_signal(&streamWriteJob.doJobCond);
    pthread_mutex_unlock(&streamWriteJob.doJobMutex);
    pthread_mutex_unlock(&streamWriteJob.workerReadyMutex);

    pthread_join(streamWriteThread, NULL);
    pthread_cond_destroy(&streamWriteJob.workerReadyCond);
    pthread_mutex_destroy(&streamWriteJob.workerReadyMutex);
    pthread_cond_destroy(&streamWriteJob.doJobCond);
    pthread_mutex_destroy(&streamWriteJob.doJobMutex);

    delete bucket;
    delete fileStream;
    delete compBucket;
}

void Writer::Flush() {
    if (bucketEvents == 0) return;

    pthread_cond_wait(&streamWriteJob.workerReadyCond, &streamWriteJob.workerReadyMutex);
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

    header->set_nevents(bucketEvents);
    header->set_bucketsize(compBucket->ByteCount());
    header->set_compression(compression);

    streamWriteJob.compBucket = compBucket;
    streamWriteJob.header = header;
    header = new proto::BucketHeader();
    streamWriteJob.isValid = true;
    pthread_mutex_lock(&streamWriteJob.doJobMutex);
    pthread_cond_signal(&streamWriteJob.doJobCond);
    pthread_mutex_unlock(&streamWriteJob.doJobMutex);

    bucket->Reset();
    bucketEvents = 0;
}

void Writer::Push(Event *event) {
    pthread_mutex_lock(&mutex);

    for (auto keyValuePair : event->metadata)
        if (metadata[keyValuePair.first] != keyValuePair.second) {
            PushMetadata(keyValuePair.first, *keyValuePair.second);
            metadata[keyValuePair.first] = keyValuePair.second;
        }

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

    pthread_mutex_unlock(&mutex);
}

void Writer::PushMetadata(std::string name, std::string &data) {
    Flush();
    (*header->mutable_metadata())[name] = data;
}

void Writer::PushMetadata(std::string name, const char *data) {
    Flush();
    (*header->mutable_metadata())[name] = data;
}

void Writer::initBucket() {
    bucket = new BucketOutputStream();
    bucketEvents = 0;
    SetCompression();
    compBucket = new BucketOutputStream();
    SetBucketDumpThreshold();
    header = new proto::BucketHeader;

    pthread_mutex_init(&streamWriteJob.doJobMutex, NULL);
    pthread_cond_init(&streamWriteJob.doJobCond, NULL);
    pthread_mutex_init(&streamWriteJob.workerReadyMutex, NULL);
    pthread_cond_init(&streamWriteJob.workerReadyCond, NULL);
    streamWriteJob.fileStream = fileStream;
    streamWriteJob.isValid = false;

    pthread_mutex_lock(&streamWriteJob.workerReadyMutex);
    pthread_create(&streamWriteThread, NULL, Writer::streamWrite, &streamWriteJob);

    pthread_mutex_init(&mutex, NULL);
}

void *Writer::streamWrite(void *writeJobVoid) {
    auto job = static_cast<WriteJob *>(writeJobVoid);

    pthread_mutex_lock(&job->doJobMutex);
    while (true) {
        pthread_mutex_lock(&job->workerReadyMutex);
        pthread_cond_signal(&job->workerReadyCond);
        pthread_mutex_unlock(&job->workerReadyMutex);
        pthread_cond_wait(&job->doJobCond, &job->doJobMutex);

        if (job->isValid) {
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
            job->isValid = false;
        } else
            break;
    }
    pthread_mutex_unlock(&job->doJobMutex);

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
