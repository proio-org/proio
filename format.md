# Format
This document describes the stream format for proio.

## Buckets
Proio streams are segmented into what are called buckets.  A bucket is a
collection of consecutive events that are optionally compressed together, and
each bucket has a header.  Buckets are intended to be O(MB)-sized, i.e. large
enough for efficient compression and also much larger than the header data.  On
disk, this also translates to bucket headers occupying a very small fraction of
the total number of disk sectors used by the proio file.  This is important for
fast direct access of events, since proio streams do not contain global
locations of events.

![proio buckets](figures/proio_buckets.png)

### Header
Each bucket has a header that describes the bucket.  This header is also an
opportunity to resynchronize/recover the stream so that in principle corruption
within a bucket is isolated.  This synchronization is achieved via a magic
number.  This is a special sequence of 128 bits that identifies the start of
the bucket header.  Following the magic number is an unsigned 32-bit
little-endian integer that states the size of the remaining header which is in
Protobuf wire format and described in [proio.proto](proio.proto).

![proio buckets](figures/bucket_header.png)

TODO: Complete this document

#### Metadata

### Contents

## Events

![proio event](figures/proio_event.png)

### Entries

### Types

### Tags

# Other sources of information
* [Pull request for major proio rewrite](https://github.com/decibelcooper/proio/pull/14)
* [EICIO: original concept that evolved into proio](https://github.com/decibelcooper/eicio)
