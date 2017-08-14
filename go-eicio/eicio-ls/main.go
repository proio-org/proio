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
		`Usage: eicio-ls [options] <eicio-input-file>
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

	var event *eicio.Event
	for event, err = reader.Next(); event != nil; event, err = reader.Next() {
		if err == eicio.ErrResync {
			log.Print(err)
		}

		fmt.Print(event)
	}

	if err != nil && err != io.EOF {
		log.Print(err)
	}
}
