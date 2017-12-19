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
