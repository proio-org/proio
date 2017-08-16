#ifndef READER_H
#define READER_H

#include "eicio.pb.h"

#include <google/protobuf/io/zero_copy_stream_impl_lite.h>

namespace eicio {
class Reader {
   public:
    Reader(int fd, bool gzip = false);
    Reader(const char *filename);
    virtual ~Reader();

    class Stream : public google::protobuf::io::CopyingInputStream {
       public:
        Stream(int fd);
        virtual ~Stream();

        int Read(void *buffer, int size);
        int Skip(int count);
    };

   private:
    Stream *stream;
};
}

#endif  // READER_H
