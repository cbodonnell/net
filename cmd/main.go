package main

import (
	"log"
	"os"

	"github.com/cbodonnell/net/cmd/commands"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <command[client|relay|server]>\n", os.Args[0])
	}

	command := os.Args[1]

	switch command {
	case "client":
		log.Fatal(commands.ClientCmd())
	case "server":
		log.Fatal(commands.ServerCmd())
	case "relay":
		log.Fatal(commands.RelayCmd())
	default:
		log.Fatalf("Unknown command: %s\n", command)
	}
}
