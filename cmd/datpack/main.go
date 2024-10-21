package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	flag.StringVar(&metadataPath, "metadatapath", "", "Path to METADATA.json file")
}

func main() {
	flag.Parse()

	if metadataPath != "" {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if err := pack(
			ctx,
			metadataPath,
			func(total, current uint32, name string) {
				log.Printf("% 8d/%d(%s): start\n", current, total, name)
			},
			func(total, current uint32, name string) {
				log.Printf("% 8d/%d(%s): done\n", current, total, name)
			},
		); err != nil {
			log.Fatalln(err)
		}
	} else {
		gui()
	}
}
