package main

import (
	"flag"
	"log"
)

func init() {
	flag.StringVar(&pngPath, "pngpath", "", "Path to PNG file")
	// TODO: format output for tim2
	// flag.StringVar(&format, "format", "", "Format output")
}

func main() {
	flag.Parse()

	if pngPath != "" {
		if err := convert(pngPath); err != nil {
			log.Fatalln(err)
		}
	} else {
		gui()
	}
}
