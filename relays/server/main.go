package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/cbodonnell/tcp-queue/pkg/crypto"
)

var relayAddress string
var serverAddress string
var key = []byte("passphrasewhichneedstobe32bytes!")

func main() {
	// var port uint
	// var certFile string
	// var keyFile string

	// flag.UintVar(&port, "relay-port", 4444, "The port to dial on the relay server")
	// flag.UintVar(&port, "server-port", 5555, "The port to dial on the server application")
	// flag.StringVar(&certFile, "tlscert", "", "TLS cert file path")
	// flag.StringVar(&keyFile, "tlskey", "", "TLS key file path")
	// flag.BoolVar(&binaryMode, "b", false, "Use binary frames instead of text frames")
	// flag.BoolVar(&binaryMode, "binary", false, "Use binary frames instead of text frames")
	flag.Usage = usage
	flag.Parse()

	relayAddress = flag.Arg(0)
	if relayAddress == "" {
		log.Fatal("No relay address specified")
	}

	serverAddress = flag.Arg(1)
	if serverAddress == "" {
		log.Fatal("No server address specified")
	}

	log.Printf("Fetching from %s\n", relayAddress)
	log.Printf("Relaying to %s\n", serverAddress)

	go func() {
		for {
			if err := fetchAndRelay(); err != nil {
				log.Println("Error relaying: ", err.Error())
				log.Println("Retrying in 1 second")
				time.Sleep(time.Second)
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <relayAddress> <serverAddress>\n", os.Args[0])
	flag.PrintDefaults()
}

func fetchAndRelay() error {
	// connect to the relay server
	log.Println("Connecting to relay server")
	relayConn, err := net.Dial("tcp", relayAddress)
	if err != nil {
		return err
	}
	defer relayConn.Close()

	// wait for a message from the relay server
	log.Println("Waiting for message from relay server")
	buf := make([]byte, 1024)
	reqLen, err := relayConn.Read(buf)
	if err != nil {
		return err
	}
	message := buf[:reqLen]

	// Decrypt the message
	message, err = crypto.Decrypt(key, message)
	if err != nil {
		return err
	}

	// connect to the server
	log.Println("Connecting to the server application")
	serverConn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return err
	}
	defer serverConn.Close()

	// send the message to the server
	log.Println("Sending message to the server application")
	_, err = serverConn.Write(message)
	if err != nil {
		return err
	}

	// wait for a response from the server
	log.Println("Waiting for a response from the server application")
	buf = make([]byte, 1024)
	reqLen, err = serverConn.Read(buf)
	if err != nil {
		return err
	}

	response := buf[:reqLen]

	// Encrypt the response
	response, err = crypto.Encrypt(key, response)
	if err != nil {
		return err
	}

	// send the response to the relay server
	log.Println("Sending response to relay server")
	_, err = relayConn.Write(response)
	if err != nil {
		return err
	}

	return nil
}
