package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cbodonnell/tcp-queue/pkg/relay"
)

func main() {
	var clientPort uint
	var serverPort uint
	var bufferSize uint
	var debug bool

	flag.UintVar(&clientPort, "client-port", 3333, "The port to listen for the client on")
	flag.UintVar(&serverPort, "server-port", 4444, "The port to listen for the server on")
	flag.UintVar(&bufferSize, "buffer-size", 1024, "The maximum size of the buffer")
	flag.BoolVar(&debug, "debug", false, "Print debug messages")
	flag.Usage = usage
	flag.Parse()

	opts := relay.RelayOpts{
		ClientPort: clientPort,
		ServerPort: serverPort,
		BufferSize: bufferSize,
		Debug:      debug,
	}
	relay := relay.NewRelay(opts)
	log.Fatal(relay.Run())
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
	flag.PrintDefaults()
}
