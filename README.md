# proio
Github: https://github.com/decibelcooper/proio

## Languages and status [![Travis CI Build Status](https://travis-ci.org/decibelcooper/proio.svg?branch=master)](https://travis-ci.org/decibelcooper/proio)
* [Go](go-proio)
  * Implemented.  See docs and issues for more information
* [Python](py-proio)
  * Implemented.  See docs and issues for more information
* [C++](cpp-proio)
  * Implemented.  See docs and issues for more information
* [Java](java-proio)
  * Code has been cleared with recent rewrite
  * Needs to be rewritten based on new scheme
  
## What is proio?
Proio is a library and set of tools that provide a simple but powerful and
performant IO for physics events.  Proio uses Google's protocol buffer
libraries and aims simply to add the concept of an event to a protocol buffer
IO.  This work was inspired and influenced by
[LCIO](https://github.com/iLCSoft/LCIO), ProMC (Sergei Chekanov), and EicMC
(Alexander Kiselev).  Another primary goal of proio is to be language-neutral,
in the sense that users can be free to use Go, Python, C++, or Java without any
significant drawback to any particular choice (proio is under active
developement, please see above for status).  Each language implementation is
written natively, and protobuf compilers generate code in each language from a
single source.  The protobuf messages described in the generated code are used
by the proio libraries to produce serialized event structures for IO.

### Events
The proio event structures can contain any protobuf messages that the user
wishes to write to the stream or file.  Each event contains a list of entries
which are the user data structures (required to be protobuf message
implementations).  This list of entries is organized with the use of tags.

![proio event](proto/figures/proio_event.png)

#### Tags
A tag is a mapping from a human-readable string to a set of event entries.
Each event carries a list of tags that point to unique entry IDs.  The concept
of tags replaces the concept of collections in LCIO.  The difference between
collections and tags is that a given tag can point to any type of data
structure, and any number of tags may point to the same entry.

#### Types
Events also carry a list of protobuf message types entered into the event by
the user.  These are string identifiers used by the protobuf libraries and
specified by the writers of the protobuf files.  Proio is distributed with
common messages that are organized into data models.  For example,
`proio.model.lcio.MCParticle` is one available type that is distributed with
proio.  Users can create and use their own types, but most users do not need to
do this, or even worry about protobuf messages at all.  The libraries
automatically determine the types of the entries and store them.  When reading
a file, the proio libraries use these type identifiers to look up message
descriptors in memory, and create objects of the appropriate type in memory to
then fill with the stored data.

#### Referencing
Following the example of LCIO, the concept of persistently referencing entries
from other entries to establish relationships is supported in proio.  For each
entry added to the event by the user, a unique ID is returned which can be
stored, e.g., by another entry.  This ID can later be used to retrieve the
desired entry from the event object.  For example, the ID of a corresponding MC
particle entry can be stored in a tracker hit entry to facilitate evaluation of
tracking software.

### Buckets
Proio writes events into what are called buckets.  A bucket is a collection of
sequential events that are compressed together, and has a header describing the
contents of the bucket (e.g. compression type and number of events contained).
The proio library adds events to a bucket until the bucket starts to overflow
based on a specified target bucket size in bytes, at which point the bucket is
written out to the stream or file.  The proio reader can use the lightweight
bucket headers to efficiently scan the contents of a file.  The bucket headers
are implemented as protobuf messages, so additional metadata can be added to
the headers without breaking compatibility.  Additionally, bucket headers can
store protobuf file descriptor protos.  These can be added to the stream by the
writer so that the reader has all of the information needed to read the user's
event data, even without having access to the particular protobuf message code
used by the user.

Compression type is set per-bucket, and can be LZ4 (default), GZIP, or
uncompressed.

![proio buckets](proto/figures/proio_buckets.png)

## Getting started
The best way to get started with proio is to look at examples.  First, pick a
language of your choice, and navigate to that subdirectory in this repository.
Follow the installation instructions given in the corresponding README.md, and
then follow some examples that are described there as well.

### Command-line tools
Most proio tools are written in go.  To fetch and build them, first make sure
you have the go compiler installed on your system, and make sure your GOPATH
and PATH environment variables are set up appropriately to effectively use `go
get`.  A typical configuration is to set `GOPATH=$HOME` and
`PATH=$GOPATH/bin:$PATH`.  Then, let's grab and install the proio package and
tools:
```shell
go get github.com/decibelcooper/proio/go-proio/...
```
#### tools
* proio-ls
* proio-summary
* proio-strip
* lcio2proio
