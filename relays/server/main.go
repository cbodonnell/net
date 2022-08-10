package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var relayAddress string
var serverAddress string

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

	responseChan := make(chan []byte)
	messageChan := make(chan []byte)

	go func(responseChan chan []byte, messageChan chan<- []byte) {
		responseChan <- []byte("init")
		defer close(messageChan)
		for {
			// wait for a response to the message
			response := <-responseChan

			// connect to the relay server
			relayConn, err := net.Dial("tcp", relayAddress)
			if err != nil {
				log.Println("dial:", err.Error())
				return
			}
			defer relayConn.Close()

			// send the response to the relay server
			_, err = relayConn.Write(response)
			if err != nil {
				log.Println("Error writing to relay:", err.Error())
			}

			// wait for a message from the relay server
			buf := make([]byte, 1024)
			reqLen, err := relayConn.Read(buf)
			if err != nil {
				log.Println("Error reading from relay:", err.Error())
			}
			message := buf[:reqLen]

			// send the message to the server
			messageChan <- message
		}
	}(responseChan, messageChan)

	go func(responseChan chan []byte, messageChan <-chan []byte) {
		defer close(responseChan)
		for {
			// get a message from the relay server
			message := <-messageChan

			// connect to the server
			serverConn, err := net.Dial("tcp", serverAddress)
			if err != nil {
				log.Println("dial:", err.Error())
				return
			}
			defer serverConn.Close()

			// send the message to the server
			_, err = serverConn.Write(message)
			if err != nil {
				log.Println("Error writing to server:", err.Error())
			}

			// wait for a response from the server
			buf := make([]byte, 1024)
			reqLen, err := serverConn.Read(buf)
			if err != nil {
				log.Println("Error reading from server:", err.Error())
			}

			response := buf[:reqLen]

			// send the response to the relay server
			responseChan <- response
		}
	}(responseChan, messageChan)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <tcpTargetAddress>\n", os.Args[0])
	flag.PrintDefaults()
}
