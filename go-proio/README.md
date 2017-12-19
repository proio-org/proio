# API
API documentation is provided by godoc.org

[![GoDoc](https://godoc.org/github.com/decibelcooper/proio/go-proio?status.svg)](https://godoc.org/github.com/decibelcooper/proio/go-proio)

# Installing
go-proio is `go get`-able.  Make sure you have go installed and `GOPATH` set up (described in [the main readme](../README.md)):
```shell
go get github.com/decibelcooper/proio/go-proio/...
```

# Examples
* [Print](example_print_test.go)
* [Scan](example_scan_test.go)
* [Skip](example_skip_test.go)
* [Push, get, inspect](example_pushGetInspect_test.go)

# Benchmarking
```shell
go test -bench=. -count 2
```
results in the following (on my little chromebook):
```
goos: linux
goarch: amd64
pkg: github.com/decibelcooper/proio/go-proio
BenchmarkWriteUncomp-4              5000            234663 ns/op
BenchmarkWriteUncomp-4              5000            212814 ns/op
BenchmarkWriteLZ4-4                 5000            512412 ns/op
BenchmarkWriteLZ4-4                 5000            482119 ns/op
BenchmarkWriteGZIP-4                5000           2172472 ns/op
BenchmarkWriteGZIP-4                5000           2121632 ns/op
BenchmarkReadUncomp-4               5000            441690 ns/op
BenchmarkReadUncomp-4               5000            444382 ns/op
BenchmarkReadLZ4-4                  5000            728634 ns/op
BenchmarkReadLZ4-4                  5000            714224 ns/op
BenchmarkReadGZIP-4                 5000            736085 ns/op
BenchmarkReadGZIP-4                 5000            740221 ns/op
PASS
ok      github.com/decibelcooper/proio/go-proio 79.248s
```
