package client

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cbodonnell/tcp-queue/pkg/client"
)

var key = []byte("passphrasewhichneedstobe32bytes!")

func ClientCmd() {
	var port uint
	var bufferSize uint
	var debug bool

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientCmd.UintVar(&port, "port", 2222, "The port to listen on")
	clientCmd.UintVar(&bufferSize, "buffer-size", 1024, "The maximum size of the buffer")
	clientCmd.BoolVar(&debug, "debug", false, "Print debug messages")
	clientCmd.Usage = func() {
		usage(clientCmd)
	}
	clientCmd.Parse(os.Args[2:])

	relayAddress := clientCmd.Arg(0)
	if relayAddress == "" {
		log.Fatal("No address specified")
	}

	client, err := client.NewClient(client.ClientOpts{
		Port:         port,
		RelayAddress: relayAddress,
		Key:          key,
		BufferSize:   bufferSize,
		Debug:        debug,
	})
	if err != nil {
		log.Fatalf("Error creating client: %s", err.Error())
	}
	log.Fatal(client.Run())
}

func usage(subcmd *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s %s [flags] <relayAddress>\n", os.Args[0], os.Args[1])
	subcmd.PrintDefaults()
}
