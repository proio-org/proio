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
BenchmarkWriteUncomp-4             10000            134603 ns/op
BenchmarkWriteUncomp-4             10000            130964 ns/op
BenchmarkWriteLZ4-4                 5000            273886 ns/op
BenchmarkWriteLZ4-4                 5000            272471 ns/op
BenchmarkWriteGZIP-4                5000           1360375 ns/op
BenchmarkWriteGZIP-4                5000           1365100 ns/op
BenchmarkReadUncomp-4               5000            279615 ns/op
BenchmarkReadUncomp-4               5000            279104 ns/op
BenchmarkReadLZ4-4                  5000            285193 ns/op
BenchmarkReadLZ4-4                  5000            285822 ns/op
BenchmarkReadGZIP-4                 5000            471545 ns/op
BenchmarkReadGZIP-4                 5000            474458 ns/op
PASS
ok      github.com/decibelcooper/proio/go-proio 50.009s
```
