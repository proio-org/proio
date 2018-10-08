# proio
Github: https://github.com/proio-org/proio

## Introduction
Proio is an event-oriented streaming data format based on Google's [protocol
buffers](https://developers.google.com/protocol-buffers/) (protobuf).  Proio
aims to add event structure and additional compression to protobuf in a way
that supports event data model serialization in medium- and high-energy
physics.  Additionally, proio
* supports self-descriptive data,
* is stream compatible,
* is language agnostic,
* and brings along many advantages of protobuf, including forward/backward
  compatibility.

For detailed information on the proio format, please see
[format.md](format.md).  This work was inspired and influenced by
[LCIO](https://github.com/iLCSoft/LCIO), ProMC (Sergei Chekanov), and EicMC
(Alexander Kiselev)

There are several language-native library implementations of proio which
support manipulating events and reading/writing streams.  Each of these
implementations adhere to the proio format, and therefore produce and consume
compatible streams.

### Language implementations
* [Go](https://github.com/proio-org/go-proio): Implemented
* [Python](https://github.com/proio-org/py-proio): Implemented
* [C++](https://github.com/proio-org/cpp-proio): Implemented
* [Java](https://github.com/proio-org/java-proio): Currently read only
  
## Getting started
The best way to get started with proio is to look at examples.  First, pick a
library in the language of your choice, and navigate to that repository.
Follow the installation instructions given in the corresponding README.md, and
then follow some examples that should be described there as well.

### Command-line tools
Most proio tools are written in go.  This is to provide highly portable but
also highly performant tools.  The tool sources are located
[here](https://github.com/proio-org/go-proio/tree/master/tools), and the tools
are `go get`-able with the following command (if you have the `go` compiler):
```shell
go get github.com/proio-org/go-proio/tools/...
```

If you do not have the `go` compiler, you can find pre-compiled binaries [in
the go-proio releases](https://github.com/proio-org/go-proio/releases).
