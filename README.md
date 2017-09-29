# What is proio?

Proio is a language-neutral IO scheme for medium- and high-energy physics.  In
order to be language-neutral, proio leverages Google's Protobuf compiler and
associated libraries in various languages.  The Protobuf compiler generates
code in each language to be used that represents the data model described in
`proio.proto`.  In addition, thin libraries have been written in each language
that are independent of the data model.  Therefore, the data model can evolve
in a language-independent way.

The name "proio" comes from the fact that it leverages Google's Protobuf, as
well as from the prefix "proto" itself.  A proto language is a common language
that others evolve from.  Similarly, proio can be the commonality between
multiple experiments with slightly different data model needs.  Each experiment
can fork proio and create necessary extensions to the core data model, and
still maintain the ability to share data through the core model.
Alternatively, experiments with particular needs can fork proio and start the
data model from scratch.  The core data model is based on
[LCIO](https://github.com/iLCSoft/LCIO).

# Data structure

Proio is a stream-oriented IO.  Data are organized into events, and a proio
stream is simply one serialized event followed by another.  A proio file
(`*.proio` or `*.proio.gz`) is no different than a proio stream, except that it
is seekable (one can look back to the beginning of the file).  As alluded to by
the possible `*.proio.gz` extension, the proio stream can be used either
compressed or uncompressed (for performance/space trade-offs), and a compressed
stream is simply an uncompressed proio stream wrapped in a gzip stream.  Proio
is designed so that the following operations are valid:
```shell
cat run1.proio run2.proio > allruns.proio
gzip run1.proio run2.proio
cat run1.proio.gz run2.proio.gz > allruns.proio.gz
```
Additionally, both uncompressed and compressed proio streams and files can be
arbitrarily large.

Within each event there is an event header - containing metadata for the event
such as a timestamp as well as internal information about the structure of the
data within the event.  The event data are stored in what is referred to as the
event payload, and within this payload is a set of collections.  Each
collection has a name and a type, and the type determines what kind of entries
can be stored in the collection.  The number of entries is arbitrary.  The
collections and entries are serialized by Protobuf, and all are described in
[proio.proto](proio.proto).  (Since the data structures are serialized by
Protobuf, integer fields gain additional compressed by Google's "varints")  As
an example of a collection type, let's look at the `MCParticle` defined in
[proio.proto](proio.proto):
```protobuf
message MCParticle {
	uint32 id = 1;
	repeated Reference parents = 2;
	repeated Reference children = 3;
	int32 PDG = 4;
	repeated double vertex = 5;
	float time = 6;
	repeated double p = 7;
	double mass = 8;
	float charge = 9;
	repeated double PEndPoint = 10;
	repeated float spin = 11;
	repeated int32 colorFlow = 12;

	int32 genStatus = 13;
	uint32 simStatus = 14;
}

message MCParticleCollection {
	uint32 id = 1;
	uint32 flags = 2;
	Params params = 3;
	repeated MCParticle entries = 4;
}
```
`MCParticle` represents generator-level particles, and these are stored in
collections of type `MCParticleCollection`.  A collection of this type may be
placed into the event payload alongside other collections with the same or
different type.  The thin language-specific libraries each provide a type
called Event which provides methods for adding, removing, and retrieving
collections from the payload.

# Getting started
## 
