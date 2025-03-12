package main

import (
	"fmt"
	"log"
	"os"

	"../pkg/xgfile"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: xgfile <input_file>")
	}

	inputFile := os.Args[1]
	importer := xgfile.NewImport(inputFile)

	segments, err := importer.GetFileSegment()
	if err != nil {
		log.Fatalf("Error extracting file segments: %v", err)
	}

	for _, segment := range segments {
		fmt.Printf("Extracted segment: %s\n", segment.Filename)
		if err := segment.Close(); err != nil {
			log.Fatalf("Error closing segment: %v", err)
		}
	}
}
