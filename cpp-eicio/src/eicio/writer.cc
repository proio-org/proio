#include <fcntl.h>
#include <string.h>
#include <iostream>

#include "event.h"
#include "writer.h"

#include <google/protobuf/io/gzip_stream.h>
#include <google/protobuf/io/zero_copy_stream_impl.h>

using namespace google::protobuf;

using namespace eicio;

Writer::Writer(int fd, bool gzip) {
    outputStream = NULL;
    fileStream = NULL;

    auto fileStream = new io::FileOutputStream(fd);
    fileStream->SetCloseOnDelete(true);
    outputStream = fileStream;

    if (gzip) {
        outputStream = new io::GzipOutputStream(outputStream);
        this->fileStream = fileStream;
    }
}

Writer::Writer(std::string filename) {
    outputStream = NULL;
    fileStream = NULL;

    int fd = open(filename.c_str(), O_WRONLY | O_CREAT | O_TRUNC, S_IRUSR | S_IWUSR | S_IRGRP | S_IROTH);
    if (fd != -1) {
        auto fileStream = new io::FileOutputStream(fd);
        fileStream->SetCloseOnDelete(true);
        outputStream = fileStream;

        string gzipSuffix = ".gz";
        int sfxLength = gzipSuffix.length();
        if (filename.length() > sfxLength) {
            if (filename.compare(filename.length() - sfxLength, sfxLength, gzipSuffix) == 0) {
                outputStream = new io::GzipOutputStream(outputStream);
                this->fileStream = fileStream;
            }
        }
    }
}

Writer::~Writer() {
    if (outputStream) delete outputStream;
    if (fileStream) delete fileStream;
}

bool Writer::Push(Event *event) {
    if (!outputStream) return false;
    auto stream = new io::CodedOutputStream(outputStream);

    event->FlushCollCache();
    unsigned int payloadSize = event->GetPayloadSize();

    for (int i = 0; i < 4; i++) stream->WriteRaw(magicBytes + i, 1);
    stream->WriteLittleEndian32((unsigned int)(event->GetHeader()->ByteSizeLong()));
    stream->WriteLittleEndian32(payloadSize);
    if (!event->GetHeader()->SerializeToCodedStream(stream)) {
        delete stream;
        return false;
    }
    stream->WriteRaw(event->GetPayload(), payloadSize);

    delete stream;
    return true;
}
