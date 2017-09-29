#ifndef WRITER_H
#define WRITER_H

#include <string.h>

#include "proio.pb.h"
#include "proio/event.h"

#include <google/protobuf/io/zero_copy_stream.h>
#include <google/protobuf/io/zero_copy_stream_impl.h>

namespace proio {
class Writer {
   public:
    Writer(int fd, bool gzip = false);
    Writer(std::string filename);
    virtual ~Writer();

    bool Push(Event *event);

   private:
    google::protobuf::io::ZeroCopyOutputStream *outputStream;
    google::protobuf::io::FileOutputStream *fileStream;
};

const unsigned char magicBytes[] = {0xe1, 0xc1, 0x00, 0x00};
}

#endif  // WRITER_H
