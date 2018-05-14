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
	outFile       = flag.String("o", "", "file to save output to")
	intersection  = flag.Bool("i", false, "only strip the intersection of the specified tags (entries that each have all tags)")
	keep          = flag.Bool("k", false, "keep only entries with the specified tags, rather than stripping them away")
	stripMetadata = flag.Bool("m", false, "strip all metadata")
	compLevel     = flag.Int("c", 1, "output compression level: 0 for uncompressed, 1 for LZ4 compression, 2 for GZIP compression")
)

func printUsage() {
	fmt.Fprintf(os.Stderr,
		`Usage: proio-strip [options] <proio-input-file> [tags...]

proio-strip will take an input proio file and either strip away entries with
specific tags, or keep only entries with specific tags.  It can also be used to
simply re-encode the proio stream by omitting tags.  By default, the output
stream is pushed to stdout, but the -o option can be used to create a file
descriptor for a specified path.

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
		writer = proio.NewWriter(os.Stdout)
	} else {
		writer, err = proio.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	switch *compLevel {
	case 2:
		writer.SetCompression(proio.GZIP)
	case 1:
		writer.SetCompression(proio.LZ4)
	default:
		writer.SetCompression(proio.UNCOMPRESSED)
	}
	defer writer.Close()

	var argTags []string
	for i := 1; i < flag.NArg(); i++ {
		argTags = append(argTags, flag.Arg(i))
	}

	nEventsRead := 0

	for event := range reader.ScanEvents() {
		if *stripMetadata {
			event.Metadata = nil
		}

		if *keep {
			keepTagIDs := make(map[uint64]bool)
			for _, keepTag := range argTags {
				for _, entryID := range event.TaggedEntries(keepTag) {
					keepTagIDs[entryID] = true
				}
			}
			for _, entryID := range event.AllEntries() {
				if !keepTagIDs[entryID] {
					event.RemoveEntry(entryID)
				}
			}
		} else {
			removeTagIDs := make(map[uint64]int)
			for _, removeTag := range argTags {
				for _, entryID := range event.TaggedEntries(removeTag) {
					removeTagIDs[entryID]++
				}
			}
			for entryID, count := range removeTagIDs {
				if !*intersection || count == len(argTags) {
					event.RemoveEntry(entryID)
				}
			}
		}

		for _, tag := range event.Tags() {
			if len(event.TaggedEntries(tag)) == 0 {
				event.DeleteTag(tag)
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
