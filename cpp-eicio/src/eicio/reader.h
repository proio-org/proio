#ifndef READER_H
#define READER_H

#include <string.h>

#include "eicio.pb.h"
#include "eicio/event.h"

#include <google/protobuf/io/zero_copy_stream.h>

namespace eicio {
class Reader {
   public:
    Reader(int fd, bool gzip = false);
    Reader(std::string filename);
    virtual ~Reader();

    Event *Get();
	int Skip(int nEvents);

   private:
    google::protobuf::uint32 syncToMagic();

    google::protobuf::io::CodedInputStream *stream;
    google::protobuf::io::ZeroCopyInputStream *inputStream;
};

const unsigned char magicBytes[] = {0xe1, 0xc1, 0x00, 0x00};
}

#endif  // READER_H
