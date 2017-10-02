package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/decibelcooper/proio/go-proio"
)

var (
	outFile    = flag.String("o", "", "file to save output to")
	keep       = flag.Bool("k", false, "keep only the specified collections, rather than stripping them away")
	decompress = flag.Bool("d", false, "decompress the stdin input with gzip")
	compress   = flag.Bool("c", false, "compress the stdout output with gzip")
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: proio-strip [options] <proio-input-file> <collections>...
options:
`,
	)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		log.Fatal("Invalid arguments")
	}

	var reader *proio.Reader
	var err error

	filename := flag.Arg(0)
	if filename == "-" {
		stdin := bufio.NewReader(os.Stdin)
		if *decompress {
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

	var writer *proio.Writer
	if *outFile == "" {
		if *compress {
			writer = proio.NewGzipWriter(os.Stdout)
		} else {
			writer = proio.NewWriter(os.Stdout)
		}
	} else {
		writer, err = proio.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer writer.Close()

	var colls []string
	for i := 1; i < flag.NArg(); i++ {
		colls = append(colls, flag.Arg(i))
	}

	nEventsRead := 0

	for event := range reader.ScanEvents() {
		for _, collName := range event.GetNames() {
			if *keep {
				keepThis := false
				for _, keepName := range colls {
					if collName == keepName {
						keepThis = true
						break
					}
				}
				if !keepThis {
					event.Remove(collName)
				}
			} else {
				for _, removeName := range colls {
					if collName == removeName {
						event.Remove(collName)
					}
				}
			}
		}

		if err := writer.Push(event); err != nil {
			log.Fatal(err)
		}

		nEventsRead++
	}

errLoop:
	for {
		select {
		case err := <-reader.Err:
			if err != io.EOF || nEventsRead == 0 {
				log.Print(err)
			}
		default:
			break errLoop
		}
	}
}
