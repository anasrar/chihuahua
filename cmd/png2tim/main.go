package main

import (
	"flag"
	"log"
)

func init() {
	flag.StringVar(&pngPath, "pngpath", "", "Path to PNG file")
	flag.UintVar(&bpp, "bpp", 8, "Bit perpixel (4 or 8)")
	// TODO: format output for tim2
	// flag.StringVar(&format, "format", "", "Format output")
}

func main() {
	flag.Parse()

	if pngPath != "" {
		if err := convert(pngPath, bpp); err != nil {
			log.Fatalln(err)
		}
	} else {
		gui()
	}
}
