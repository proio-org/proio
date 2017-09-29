# What is proio?
Proio is a language-neutral IO scheme for medium- and high-energy physics.  It was
born from an exploratory project at Argonne National Laboratory called
[eicio](https://github.com/decibelcooper/eicio).  In
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

## Referencing
Another key feature of proio is referencing.  This is a concept taken from
[LCIO](https://github.com/iLCSoft/LCIO) where persistent references to any
collection or collection entry can be stored in another.  This is implemented
in the thin language-specific libraries with Reference() and Dereference()
methods.  These methods are associated with the event, which keeps track of the
number of unique identifiers.  The Reference() methods will take a collection
or entry as an argument and return a Reference type object, which is a message
described in [proio.proto](proio.proto).  Conversely, Dereference() will take a
Reference type object and return the collection or entry that the Reference
points to.  One example of this can be found at the top of the go-proio API
documentation:
[![GoDoc](https://godoc.org/github.com/decibelcooper/proio/go-proio?status.svg)](https://godoc.org/github.com/decibelcooper/proio/go-proio)

# Getting started
## Command-line tools
Tools are specific to each language implementation.  However, most tools will
continue to be written in go, and there are already a few useful ones.  Let's
try out a few.  First, make sure you have the go compiler installed on your
system, and make sure your GOPATH and PATH environment variables are set up
appropriately to effectively use `go get`.  A typical configuration is to set
`GOPATH=$HOME` and `PATH=$GOPATH/bin:$PATH`.  Then, let's grab and install the
proio package:
```shell
go get github.com/decibelcooper/proio/go-proio/...
```
And then we can play around with manipulating the [sample files](samples/)...
### Summarize and dump
```shell
proio-summary samples/smallSample.proio
proio-ls samples/smallSample.proio | less
```
### Strip
```shell
proio-strip samples/smallSample.proio MCParticle BeamCalHits | proio-summary -
proio-strip -k samples/smallSample.proio MCParticle BeamCalHits | proio-summary -
```
### Concat
```shell
cp samples/smallSample.proio.gz tmp.proio.gz
cat samples/smallSample.proio.gz tmp.proio.gz | proio-summary -g -
```
### Cut
```shell
dd if=samples/smallSample.proio of=roughCut.proio bs=500K count=1
proio-strip -o cleanCut.proio.gz roughCut.proio
proio-summary cleanCut.proio.gz
```

## Read examples
### Go
```go
package main
  
import (
    "fmt"
    "log"

    "github.com/decibelcooper/proio/go-proio"
    "github.com/decibelcooper/proio/go-proio/model"
)

func main() {
    reader, err := proio.Open("samples/smallSample.proio.gz")
    if err != nil {
        log.Fatal(err)
    }

    for event := range reader.Events() {
        if reader.Err != nil {
            log.Println(reader.Err)
        }

        mcColl := event.Get("MCParticle").(*model.MCParticleCollection)
        if mcColl == nil || len(mcColl.Entries) < 1 {
            continue
        }

        mcPart := mcColl.Entries[0]
        fmt.Println(mcPart)
    }
}
```
### Python
```python
import proio
  
with proio.Reader("samples/smallSample.proio.gz") as reader:
    for event in reader:
        mc_coll = event.get("MCParticle")
        if mc_coll == None or len(mc_coll.entries) < 1:
            continue

        mc_part = mc_coll.entries[0]
        print(mc_part)
```
### C++
```cpp
#include <iostream>

#include "proio/event.h"
#include "proio/proio.pb.h"
#include "proio/reader.h"

int main(int argc, const char **argv) {
    auto reader = new proio::Reader("samples/smallSample.proio.gz");

    proio::Event *event;
    while ((event = reader->Get()) != NULL) {
        auto mcColl = (proio::model::MCParticleCollection *)event->Get("MCParticle");
        if (mcColl == NULL || mcColl->entries_size() < 1) continue;

        proio::model::MCParticle mcPart = mcColl->entries(0);
        std::cout << mcPart.DebugString() << std::endl;

		delete event;
    }

    delete reader;
    return EXIT_SUCCESS;
}
```
### Java
```java
import proio.Event;
import proio.Model;
import proio.Reader;

public class Read
{
    public static void main( String[] args )
    {
        try {
            Reader reader = new Reader("samples/smallSample.proio.gz");
            if (reader == null) return;

            for (Event event : reader) {
                Model.MCParticleCollection mcColl = (Model.MCParticleCollection) event.get("MCParticle");
                if (mcColl == null || mcColl.getEntriesCount() < 1) continue;

                Model.MCParticle mcPart = mcColl.getEntries(0);
                System.out.println(mcPart);
            }

            reader.close();
        } catch (Throwable e) {
            e.printStackTrace();
        }
    }
}
```

## Write examples
### Go
```go
package main

import (
    "log"

    "github.com/decibelcooper/proio/go-proio"
    "github.com/decibelcooper/proio/go-proio/model"
)

func main() {
    writer, err := proio.Create("test.proio.gz")
    if err != nil {
        log.Fatal(err)
    }
    defer writer.Close()

    event := proio.NewEvent()
    mcColl := &model.MCParticleCollection{}
    event.Add(mcColl, "MCParticle")

    parent := &model.MCParticle{}
    parent.PDG = 443
    mcColl.Entries = append(mcColl.Entries, parent)

    child1 := &model.MCParticle{}
    child1.PDG = 11
    child2 := &model.MCParticle{}
    child2.PDG = -11
    mcColl.Entries = append(mcColl.Entries, child1, child2)

    parent.Children = append(parent.Children, event.Reference(child1), event.Reference(child2))
    child1.Parents = append(child1.Children, event.Reference(parent))
    child2.Parents = append(child2.Children, event.Reference(parent))

    writer.Push(event)
}
```
### Python
```python
import proio
import proio.model as model

with proio.Writer("test.proio.gz") as writer:
    event = proio.Event()
    mc_coll = model.MCParticleCollection()
    event.add(mc_coll, "MCParticle")

    parent = mc_coll.entries.add()
    parent.PDG = 443

    child1 = mc_coll.entries.add()
    child1.PDG = 11
    child2 = mc_coll.entries.add()
    child2.PDG = -11

    parent.children.extend([event.reference(child1), event.reference(child2)])
    child1.parents.extend([event.reference(parent)])
    child2.parents.extend([event.reference(parent)])

    writer.push(event)
```
### C++
```cpp
#include <iostream>

#include "proio/event.h"
#include "proio/proio.pb.h"
#include "proio/writer.h"

int main(int argc, const char **argv) {
    auto writer = new proio::Writer("test.proio.gz");

    auto event = new proio::Event();
    auto mcColl = new proio::model::MCParticleCollection();
    event->Add(mcColl, "MCParticle");

    proio::model::MCParticle *parent = mcColl->add_entries();
    parent->set_pdg(443);

    proio::model::MCParticle *child1 = mcColl->add_entries();
    child1->set_pdg(11);
    proio::model::MCParticle *child2 = mcColl->add_entries();
    child2->set_pdg(-11);

    event->Reference(child1, parent->add_children());
    event->Reference(child2, parent->add_children());
    event->Reference(parent, child1->add_parents());
    event->Reference(parent, child2->add_parents());

    writer->Push(event);

    delete event;

    delete writer;
    return EXIT_SUCCESS;
}
```

# Modifying the data model
The data model is described in the [proio.proto](proio.proto) files.  At the
top of the file is the description of some messages that are needed for the
thin libraries.  Scroll down until you see the comment `DATA MODEL MESSAGES`.
Anything below this comment can be modified within a few simple rules:
1. For message type Msg, there must be a corresponding collection type named
   MsgCollection.
2. Every message type and collection type must have a `uint32 id` field
   assigned any number.
3. For collection type MsgCollection, there must be a `repeated Msg entries`
   field assigned any number.
Other than the above rules, anything goes.  Any number of message and
collection types may be defined.  Please see the [Protobuf Language
Guide](https://developers.google.com/protocol-buffers/docs/proto3) for details
on the syntax.

## Generating the code after modifying the model
Any time [proio.proto](proio.proto) is modified, the language-specific code
that describes the data model must be regenerated.  For consistency, and
because there are four different languages to generate, the code generation is
done inside a container.  In particular, we use a
[Singularity](http://singularity.lbl.gov/) container.  In order to generate the
code, Singularity 2.3+ must be installed on your computer.  Once it is
installed, simply:
```shell
make
```
This will download over 1 GiB of container image layers from Docker Hub, and
place the layers into a single image file called `proio-gen.img` that is twice
as large (i.e. over 2 GiB).  Please make sure you have the space before calling
`make`.  Once the image is created, a container will automatically be run to
generate each piece of data model code for the different languages.

## Backwards and forwards compatibility
It is important to understand a little bit about how Protobuf works to
understand what changes can be made to the data model that maintain
compatibility with other versions of the code.  Please see the Protobuf
documentation for details, but here are some important things to note:
* New fields can be added without breaking forwards or backwards compatibility
  as long as new unique field numbers are used.  This is because Protobuf will
  ignore unknown fields, and (staring with Protobuf 3) all fields are optional.
* Fields can be removed from the data model for the same reason.  It is
  important, however, to never reuse the field number or name of the removed
  field.  In order to enforce this, the [reserved
  tag](https://developers.google.com/protocol-buffers/docs/proto3#reserved)
  should be use.
* For the space-conscious, field numbers should be used wisely.  The reason is
  that field numbers between 1 and 15 take one byte to identify, while larger
  field numbers must be identified with at least two bytes.
