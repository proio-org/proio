# proio for Go
## API
API documentation is provided by godoc.org

[![GoDoc](https://godoc.org/github.com/decibelcooper/proio/go-proio?status.svg)](https://godoc.org/github.com/decibelcooper/proio/go-proio)

## Installation
go-proio is `go get`-able.  Make sure you have go installed and `GOPATH` set up (described in [the main readme](../README.md)):
```shell
go get github.com/decibelcooper/proio/go-proio/...
```

For information on what versions of Go are supported, please see the [Travis CI page](https://travis-ci.org/decibelcooper/proio).

## Examples
* [Print](example_print_test.go)
* [Scan](example_scan_test.go)
* [Skip](example_skip_test.go)
* [Push, get, inspect](example_pushGetInspect_test.go)

## Benchmarking
```shell
go test -v -run=^$ -bench=. -count=2
```
results in the following (on my Core i5 desktop):
```
goos: linux
goarch: amd64
pkg: github.com/decibelcooper/proio/go-proio
BenchmarkWriteUncomp-4                     10000            128412 ns/op
BenchmarkWriteUncomp-4                     10000            127741 ns/op
BenchmarkWriteLZ4-4                         5000            267338 ns/op
BenchmarkWriteLZ4-4                         5000            268786 ns/op
BenchmarkWriteGZIP-4                        5000           1348183 ns/op
BenchmarkWriteGZIP-4                        5000           1357786 ns/op
BenchmarkReadUncomp-4                       5000            277247 ns/op
BenchmarkReadUncomp-4                       5000            276152 ns/op
BenchmarkReadLZ4-4                          5000            285991 ns/op
BenchmarkReadLZ4-4                          5000            282309 ns/op
BenchmarkReadGZIP-4                         5000            476960 ns/op
BenchmarkReadGZIP-4                         5000            477229 ns/op
BenchmarkAddRemove100Entries-4             30000             71958 ns/op
BenchmarkAddRemove100Entries-4             20000             63136 ns/op
BenchmarkAddRemove1000Entries-4             3000            742337 ns/op
BenchmarkAddRemove1000Entries-4             2000            794221 ns/op
BenchmarkAddRemove10000Entries-4             200           7657906 ns/op
BenchmarkAddRemove10000Entries-4             200           8777519 ns/op
BenchmarkAddRemove100000Entries-4             20          66231187 ns/op
BenchmarkAddRemove100000Entries-4             20          64927346 ns/op
PASS
ok      github.com/decibelcooper/proio/go-proio 65.545s
```
