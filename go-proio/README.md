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

# Benchmarks
With files:
```
-rw-r--r-- 1 dblyth dblyth 2.8G Dec  4 19:14 repeatedSampleGZIP.proio
-rw-r--r-- 1 dblyth dblyth 3.8G Dec  4 19:01 repeatedSample.proio
-rw-r--r-- 1 dblyth dblyth 5.2G Dec  4 19:13 repeatedSampleUncomp.proio
-rw-r--r-- 1 dblyth dblyth 22M Oct  2 10:35 ../samples/largeSample.slcio
```
and after clearning the kernel page cache, I get
```
BenchmarkTracking-4       	   50000	    767342 ns/op
BenchmarkTrackingLZ4-4    	   50000	    913342 ns/op
BenchmarkTrackingGzip-4   	   50000	   1810212 ns/op
BenchmarkTrackingLCIO-4   	     100	  22378964 ns/op
```

As a benchmark of the ability to scan the file headers, for
```shell
time proio-summary repeatedSample.proio
```
I get (after clearing the kernel page cache again)
```
Number of events: 50000

real	0m0.035s
user	0m0.009s
sys	0m0.000s
```
