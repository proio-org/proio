#include <fcntl.h>
#include <string.h>
#include <iostream>

#include "event.h"
#include "reader.h"
#include "writer.h"

#include <google/protobuf/io/gzip_stream.h>
#include <google/protobuf/io/zero_copy_stream_impl.h>

using namespace google::protobuf;

using namespace eicio;

Reader::Reader(int fd, bool gzip) {
    inputStream = NULL;
    fileStream = NULL;

    auto fileStream = new io::FileInputStream(fd);
    fileStream->SetCloseOnDelete(true);
    inputStream = fileStream;

    if (gzip) {
        inputStream = new io::GzipInputStream(inputStream);
        this->fileStream = fileStream;
    }
}

Reader::Reader(std::string filename) {
    inputStream = NULL;
    fileStream = NULL;

    int fd = open(filename.c_str(), O_RDONLY);
    if (fd != -1) {
        auto fileStream = new io::FileInputStream(fd);
        fileStream->SetCloseOnDelete(true);
        inputStream = fileStream;

        string gzipSuffix = ".gz";
        int sfxLength = gzipSuffix.length();
        if (filename.length() > sfxLength) {
            if (filename.compare(filename.length() - sfxLength, sfxLength, gzipSuffix) == 0) {
                inputStream = new io::GzipInputStream(inputStream);
                this->fileStream = fileStream;
            }
        }
    }
}

Reader::~Reader() {
    if (inputStream) delete inputStream;
    if (fileStream) delete fileStream;
}

Event *Reader::Get() {  // TODO: figure out error handling for this
    if (!inputStream) return NULL;
    io::CodedInputStream stream(inputStream);

    uint32 n;
    if ((n = syncToMagic(&stream)) < 4) return NULL;

    uint32 headerSize;
    if (!stream.ReadLittleEndian32(&headerSize)) return NULL;
    uint32 payloadSize;
    if (!stream.ReadLittleEndian32(&payloadSize)) return NULL;

    auto headerLimit = stream.PushLimit(headerSize);
    auto header = new model::EventHeader;
    if (!header->MergeFromCodedStream(&stream) || !stream.ConsumedEntireMessage()) {
        delete header;
        return Get();  // Indefinitely attempt to resync to magic numbers
    }
    stream.PopLimit(headerLimit);

    auto event = new Event;
    event->SetHeader(header);
    auto *payload = (unsigned char *)event->SetPayloadSize(payloadSize);
    if (!stream.ReadRaw(payload, payloadSize)) {
        delete event;
        return NULL;
    }

    return event;
}

model::EventHeader *Reader::GetHeader() {  // TODO: figure out error handling for this
    if (!inputStream) return NULL;
    io::CodedInputStream stream(inputStream);

    uint32 n;
    if ((n = syncToMagic(&stream)) < 4) return NULL;

    uint32 headerSize;
    if (!stream.ReadLittleEndian32(&headerSize)) return NULL;
    uint32 payloadSize;
    if (!stream.ReadLittleEndian32(&payloadSize)) return NULL;

    auto headerLimit = stream.PushLimit(headerSize);
    auto header = new model::EventHeader;
    if (!header->MergeFromCodedStream(&stream) || !stream.ConsumedEntireMessage()) {
        delete header;
        return GetHeader();  // Indefinitely attempt to resync to magic numbers
    }
    stream.PopLimit(headerLimit);

    if (!stream.Skip(payloadSize)) {
        delete header;
        return NULL;
    }

    return header;
}

int Reader::Skip(int nEvents) {
    if (!inputStream) return -1;
    io::CodedInputStream stream(inputStream);

    int nSkipped = 0;
    for (int i = 0; i < nEvents; i++) {
        uint32 n;
        if ((n = syncToMagic(&stream)) < 4) return -1;

        uint32 headerSize;
        if (!stream.ReadLittleEndian32(&headerSize)) return -1;
        uint32 payloadSize;
        if (!stream.ReadLittleEndian32(&payloadSize)) return -1;

        if (!stream.Skip(headerSize + payloadSize)) return -1;

        nSkipped++;
    }

    return nSkipped;
}

uint32 Reader::syncToMagic(io::CodedInputStream *stream) {
    unsigned char num;
    uint32 nRead = 0;

    while (stream->ReadRaw(&num, 1)) {
        nRead++;

        if (num == magicBytes[0]) {
            bool goodSeq = true;

            for (int i = 1; i < 4; i++) {
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
