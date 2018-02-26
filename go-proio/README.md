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
BenchmarkWriteUncomp-4                     10000            139903 ns/op
BenchmarkWriteUncomp-4                     10000            122604 ns/op
BenchmarkWriteLZ4-4                         5000            275437 ns/op
BenchmarkWriteLZ4-4                         5000            278510 ns/op
BenchmarkWriteGZIP-4                        5000           1364511 ns/op
BenchmarkWriteGZIP-4                        5000           1365314 ns/op
BenchmarkReadUncomp-4                       5000            254130 ns/op
BenchmarkReadUncomp-4                       5000            244430 ns/op
BenchmarkReadLZ4-4                          5000            249543 ns/op
BenchmarkReadLZ4-4                          5000            247739 ns/op
BenchmarkReadGZIP-4                         5000            434379 ns/op
BenchmarkReadGZIP-4                         5000            437608 ns/op
BenchmarkAddRemove100Entries-4             30000             65783 ns/op
BenchmarkAddRemove100Entries-4             30000             51931 ns/op
BenchmarkAddRemove1000Entries-4             3000            612425 ns/op
BenchmarkAddRemove1000Entries-4             3000            541261 ns/op
BenchmarkAddRemove10000Entries-4             200           6023315 ns/op
BenchmarkAddRemove10000Entries-4             300           6985336 ns/op
BenchmarkAddRemove100000Entries-4             20          61059149 ns/op
BenchmarkAddRemove100000Entries-4             20          61536698 ns/op
PASS
ok      github.com/decibelcooper/proio/go-proio 63.981s
```
