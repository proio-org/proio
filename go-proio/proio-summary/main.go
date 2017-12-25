package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/proto"
)

var ()

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: proio-summary [options] <proio-input-file>
options:
`,
	)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() != 1 {
		printUsage()
		log.Fatal("Invalid arguments")
	}

	var reader *proio.Reader
	var err error

	filename := flag.Arg(0)
	if filename == "-" {
		stdin := bufio.NewReader(os.Stdin)
		reader = proio.NewReader(stdin)
	} else {
		reader, err = proio.Open(filename)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	nBuckets := make(map[proto.BucketHeader_CompType]int)
	nEvents := 0

	var header *proto.BucketHeader
	for header, err = reader.NextHeader(); header != nil; header, err = reader.NextHeader() {
		if err != nil {
			log.Print(err)
		}

		nBuckets[header.Compression]++
		nEvents += int(header.NEvents)
	}

	if err != nil && err != io.EOF {
		log.Print(err)
	}

	fmt.Println("Number of LZ4 buckets:", nBuckets[proto.BucketHeader_LZ4])
	fmt.Println("Number of GZIP buckets:", nBuckets[proto.BucketHeader_GZIP])
	fmt.Println("Number of uncompressed buckets:", nBuckets[proto.BucketHeader_NONE])
	fmt.Println("Number of events:", nEvents)
}
