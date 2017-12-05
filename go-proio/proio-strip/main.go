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
	outFile   = flag.String("o", "", "file to save output to")
	keep      = flag.Bool("k", false, "keep only entries with the specified tags, rather than stripping them away")
	compLevel = flag.Int("c", 1, "compression level: 0 for uncompressed, 1 for LZ4 compression, 2 for GZIP compression")
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: proio-strip [options] <proio-input-file> <tags>...
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
		reader = proio.NewReader(stdin)
	} else {
		reader, err = proio.Open(filename)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	var writer *proio.Writer
	if *outFile == "" {
		switch *compLevel {
		case 2:
			writer = proio.NewWriter(os.Stdout, proio.GZIP)
		case 1:
			writer = proio.NewWriter(os.Stdout, proio.LZ4)
		default:
			writer = proio.NewWriter(os.Stdout, proio.UNCOMPRESSED)
		}
	} else {
		switch *compLevel {
		case 2:
			writer, err = proio.Create(*outFile, proio.GZIP)
		case 1:
			writer, err = proio.Create(*outFile, proio.LZ4)
		default:
			writer, err = proio.Create(*outFile, proio.UNCOMPRESSED)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	defer writer.Close()

	var argTags []string
	for i := 1; i < flag.NArg(); i++ {
		argTags = append(argTags, flag.Arg(i))
	}

	nEventsRead := 0

	for event := range reader.ScanEvents() {
		if *keep {
			keepTagIDs := make(map[uint64]bool)
			keepTags := make(map[string]bool)
			for _, keepTag := range argTags {
				keepTags[keepTag] = true
				for _, entryID := range event.TaggedEntries(keepTag) {
					keepTagIDs[entryID] = true
				}
			}
			for _, entryID := range event.AllEntries() {
				if !keepTagIDs[entryID] {
					event.RemoveEntry(entryID)
				}
			}
			for _, tag := range event.Tags() {
				if !keepTags[tag] {
					event.RemoveTag(tag)
				}
			}
		} else {
			for _, removeTag := range argTags {
				for _, entryID := range event.TaggedEntries(removeTag) {
					event.RemoveEntry(entryID)
				}
				event.RemoveTag(removeTag)
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
