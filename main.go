package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jpeach/wotcher/pkg/cli"
)

const Progname = "wotcher"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	if err := cli.NewWatcher().Execute(); err != nil {
		if msg := err.Error(); msg != "" {
			fmt.Fprintf(os.Stderr, "%s: %s\n", Progname, msg)
		}

		os.Exit(1)
	}
}
