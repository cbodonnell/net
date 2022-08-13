package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cbodonnell/tcp-queue/pkg/server"
)

var key = []byte("passphrasewhichneedstobe32bytes!")

func main() {
	var retryDuration string
	var debug bool
	flag.StringVar(&retryDuration, "retry-duration", "1s", "The duration to wait between retries")
	flag.BoolVar(&debug, "debug", false, "Print debug messages")
	flag.Usage = usage
	flag.Parse()

	relayAddress := flag.Arg(0)
	if relayAddress == "" {
		log.Fatal("No relay address specified")
	}

	serverAddress := flag.Arg(1)
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

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <relayAddress> <serverAddress>\n", os.Args[0])
	flag.PrintDefaults()
}
