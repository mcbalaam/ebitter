package main

import (
	"flag"
	"log"

	"github.com/mcbalaam/ebitter/internal/assets"
)

func main() {
	fonts := flag.String("fonts", "", "directory containing TTF fonts")
	output := flag.String("output", "", "directory for generated sprite sheets")
	flag.Parse()
	if *fonts == "" || *output == "" {
		log.Fatal("both -fonts and -output are required")
	}
	if err := assets.ProcessFonts(*fonts, *output); err != nil {
		log.Fatal(err)
	}
}
