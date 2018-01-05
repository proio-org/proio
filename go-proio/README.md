# proio for Go
## API
API documentation is provided by godoc.org

[![GoDoc](https://godoc.org/github.com/decibelcooper/proio/go-proio?status.svg)](https://godoc.org/github.com/decibelcooper/proio/go-proio)

## Installing
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
BenchmarkWriteUncomp-4             10000            131552 ns/op
BenchmarkWriteUncomp-4             10000            125353 ns/op
BenchmarkWriteLZ4-4                 5000            201523 ns/op
BenchmarkWriteLZ4-4                10000            194178 ns/op
BenchmarkWriteGZIP-4                5000           1357492 ns/op
BenchmarkWriteGZIP-4                5000           1369524 ns/op
BenchmarkReadUncomp-4               5000            275899 ns/op
BenchmarkReadUncomp-4               5000            276390 ns/op
BenchmarkReadLZ4-4                  5000            288780 ns/op
BenchmarkReadLZ4-4                  5000            295727 ns/op
BenchmarkReadGZIP-4                 5000            469517 ns/op
BenchmarkReadGZIP-4                 5000            465547 ns/op
PASS
ok      github.com/decibelcooper/proio/go-proio 50.934s
```
