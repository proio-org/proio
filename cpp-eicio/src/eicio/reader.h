#ifndef READER_H
#define READER_H

#include <string.h>

#include "eicio.pb.h"
#include "eicio/event.h"

#include <google/protobuf/io/zero_copy_stream.h>
#include <google/protobuf/io/zero_copy_stream_impl.h>

namespace eicio {
class Reader {
   public:
    Reader(int fd, bool gzip = false);
    Reader(std::string filename);
    virtual ~Reader();

    Event *Get();
    model::EventHeader *GetHeader();
    int Skip(int nEvents);

   private:
    google::protobuf::uint32 syncToMagic(google::protobuf::io::CodedInputStream *stream);

    google::protobuf::io::ZeroCopyInputStream *inputStream;
    google::protobuf::io::FileInputStream *fileStream;
};
}

#endif  // READER_H
