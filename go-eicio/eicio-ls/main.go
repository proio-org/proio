package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/decibelcooper/eicio/go-eicio"
)

var (
	doGzip = flag.Bool("g", false, "decompress the stdin input with gzip")
	event  = flag.Int("e", -1, "list specified event, numbered consecutively from the start of the file or stream")
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
		stdin := bufio.NewReader(os.Stdin)
		if *doGzip {
			reader, err = eicio.NewGzipReader(stdin)
		} else {
			reader = eicio.NewReader(stdin)
		}
	} else {
		reader, err = eicio.Open(filename)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	singleEvent := false
	if *event >= 0 {
		singleEvent = true
		_, err = reader.Skip(*event)
		if err == eicio.ErrResync {
			log.Print(err)
		} else if err != nil {
			log.Fatal(err)
		}
	}

	nEventsRead := 0

	for event := range reader.Events() {
		if reader.Err != nil {
			log.Print(reader.Err)
		}

		fmt.Print(event)

		nEventsRead++
		if singleEvent {
			break
		}
	}

	if (reader.Err != nil && reader.Err != io.EOF) || nEventsRead == 0 {
		log.Print(reader.Err)
	}
}
