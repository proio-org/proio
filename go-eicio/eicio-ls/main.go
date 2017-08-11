package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/decibelCooper/eicio/go-eicio"
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

	reader, err := eicio.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	for {
		event, _ := reader.GetEvent()
		if event == nil {
			break
		}

		fmt.Print(event)
	}
}
