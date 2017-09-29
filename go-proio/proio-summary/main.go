package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/decibelcooper/proio/go-proio"
	"github.com/decibelcooper/proio/go-proio/model"
	humanize "github.com/dustin/go-humanize"
)

var (
	doGzip = flag.Bool("g", false, "decompress the stdin input with gzip")
)

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
		if *doGzip {
			reader, err = proio.NewGzipReader(stdin)
		} else {
			reader = proio.NewReader(stdin)
		}
	} else {
		reader, err = proio.Open(filename)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	nEvents := 0
	colls := make([]string, 0)
	collBytes := make(map[string]uint64)
	runs := make(map[uint64]bool)

	var header *model.EventHeader
	for header, err = reader.GetHeader(); header != nil; header, err = reader.GetHeader() {
		if err != nil {
			log.Print(err)
		}

		runs[header.RunNumber] = true
		nEvents++

		for _, collHdr := range header.PayloadCollections {
			if _, ok := collBytes[collHdr.Type]; !ok {
				colls = append(colls, collHdr.Type)
				sort.Strings(colls)
			}
			collBytes[collHdr.Type] += uint64(collHdr.PayloadSize)
		}
	}

	if err != nil && err != io.EOF {
		log.Print(err)
	}

	fmt.Println("Number of runs:", len(runs))
	fmt.Println("Number of events:", nEvents)
	fmt.Println("Total bytes for...")
	for _, key := range colls {
		fmt.Print("\t", key, ": ", humanize.Bytes(collBytes[key]), "\n")
	}
}
