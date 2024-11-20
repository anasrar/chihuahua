package main

import (
	"flag"
	"log"
)

func init() {
	flag.StringVar(&pngPath, "pngpath", "", "Path to PNG file")
	flag.UintVar(&bpp, "bpp", 8, "Bit perpixel (4 or 8)")
	flag.StringVar(&format, "format", "TIM3", "Format output")
}

func main() {
	flag.Parse()

	if pngPath != "" {
		switch format {
		case "TIM3":
		case "TIM2":
		default:
			log.Fatalln("Allowed format is TIM3 and TIM2")
		}

		if err := convert(pngPath, bpp, format); err != nil {
			log.Fatalln(err)
		}
	} else {
		gui()
	}
}
