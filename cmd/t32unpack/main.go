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
	flag.StringVar(&t32Path, "t32path", "", "Path to t32 file")
}

func main() {
	flag.Parse()

	if t32Path != "" {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if err := unpack(
			ctx,
			t32Path,
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
