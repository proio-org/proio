package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/decibelCooper/eicio/go-eicio"
)

var (
	doGzip = flag.Bool("g", false, "decompress the stdin input with gzip")
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: eicio-summary [options] <eicio-input-file>
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

	var reader *eicio.Reader
	var err error

	filename := flag.Arg(0)
	if filename == "-" {
		if *doGzip {
			reader, err = eicio.NewGzipReader(os.Stdin)
		} else {
			reader = eicio.NewReader(os.Stdin)
		}
	} else {
		reader, err = eicio.Open(filename)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	nEvents := 0

	var header *eicio.EventHeader
	for header, err = reader.NextHeader(); header != nil; header, err = reader.NextHeader() {
		if err != nil {
			log.Print(err)
		}

		nEvents++
	}

	if err != nil && err != io.EOF {
		log.Print(err)
	}

	fmt.Println("Number of events:", nEvents)
}
