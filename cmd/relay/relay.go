package relay

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cbodonnell/tcp-queue/pkg/relay"
)

func RelayCmd() {
	var clientPort uint
	var serverPort uint
	var bufferSize uint
	var debug bool

	relayCmd := flag.NewFlagSet("relay", flag.ExitOnError)
	relayCmd.UintVar(&clientPort, "client-port", 3333, "The port to listen for the client on")
	relayCmd.UintVar(&serverPort, "server-port", 4444, "The port to listen for the server on")
	relayCmd.UintVar(&bufferSize, "buffer-size", 1024, "The maximum size of the buffer")
	relayCmd.BoolVar(&debug, "debug", false, "Print debug messages")
	relayCmd.Usage = func() {
		usage(relayCmd)
	}
	relayCmd.Parse(os.Args[2:])

	opts := relay.RelayOpts{
		ClientPort: clientPort,
		ServerPort: serverPort,
		BufferSize: bufferSize,
		Debug:      debug,
	}
	relay := relay.NewRelay(opts)
	log.Fatal(relay.Run())
}

func usage(subcmd *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s %s [flags]\n", os.Args[0], os.Args[1])
	subcmd.PrintDefaults()
}
