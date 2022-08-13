package main

import (
	"log"
	"os"

	"github.com/cbodonnell/tcp-queue/cmd/client"
	"github.com/cbodonnell/tcp-queue/cmd/relay"
	"github.com/cbodonnell/tcp-queue/cmd/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <command>\n", os.Args[0])
	}

	switch os.Args[1] {
	case "client":
		client.ClientCmd()
	case "server":
		server.ServerCmd()
	case "relay":
		relay.RelayCmd()
	default:
		log.Fatalf("Unknown command: %s\n", os.Args[1])
	}
}
