package server

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cbodonnell/tcp-queue/pkg/server"
)

var key = []byte("passphrasewhichneedstobe32bytes!")

func ServerCmd() {
	var retryDuration string
	var debug bool

	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverCmd.StringVar(&retryDuration, "retry-duration", "1s", "The duration to wait between retries")
	serverCmd.BoolVar(&debug, "debug", false, "Print debug messages")
	serverCmd.Usage = func() {
		usage(serverCmd)
	}
	serverCmd.Parse(os.Args[2:])

	relayAddress := serverCmd.Arg(0)
	if relayAddress == "" {
		log.Fatal("No relay address specified")
	}

	serverAddress := serverCmd.Arg(1)
	if serverAddress == "" {
		log.Fatal("No server address specified")
	}

	server, err := server.NewServer(server.ServerOpts{
		RelayAddress:  relayAddress,
		ServerAddress: serverAddress,
		Key:           key,
		RetryDuration: retryDuration,
		Debug:         debug,
	})
	if err != nil {
		log.Fatalf("Error creating server: %s", err.Error())
	}
	log.Fatal(server.Run())
}

func usage(subcmd *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s %s [flags] <relayAddress> <serverAddress>\n", os.Args[0], os.Args[1])
	subcmd.PrintDefaults()
}
