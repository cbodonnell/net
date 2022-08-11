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

	log.Printf("Fetching from %s\n", relayAddress)
	log.Printf("Relaying to %s\n", serverAddress)

	done := make(chan bool)

	responseChan := make(chan []byte)
	messageChan := make(chan []byte)

	go func(responseChan chan []byte, messageChan chan<- []byte) {
		response := []byte("init")
		defer close(messageChan)
		for {
			// connect to the relay server
			log.Println("Connecting to relay server")
			relayConn, err := net.Dial("tcp", relayAddress)
			if err != nil {
				log.Println("dial:", err.Error())
				return
			}
			defer relayConn.Close()

			// wait for a message from the relay server
			log.Println("Waiting for message from relay server")
			buf := make([]byte, 1024)
			reqLen, err := relayConn.Read(buf)
			if err != nil {
				log.Println("Error reading from relay:", err.Error())
			}
			message := buf[:reqLen]

			// send the message to the server
			log.Println("Sending message to server application")
			messageChan <- message

			// wait for a response to the message
			log.Println("waiting for response from server application")
			response = <-responseChan
			log.Printf("Received response with %d bytes:\n%s\n", len(response), string(response))

			// send the response to the relay server
			log.Println("Sending response to relay server")
			_, err = relayConn.Write(response)
			if err != nil {
				log.Println("Error writing to relay:", err.Error())
			}
		}
	}(responseChan, messageChan)

	go func(responseChan chan []byte, messageChan <-chan []byte) {
		defer close(responseChan)
		for {
			// get a message from the relay server
			log.Println("TO_SERVER: Waiting for message for server application")
			message := <-messageChan
			log.Printf("Received message with %d bytes:\n%s\n", len(message), string(message))

			// connect to the server

			log.Println("TO_SERVER: Connecting to the server application")
			serverConn, err := net.Dial("tcp", serverAddress)
			if err != nil {
				log.Println("dial:", err.Error())
				return
			}
			defer serverConn.Close()

			// send the message to the server
			log.Println("TO_SERVER: Sending message to the server application")
			_, err = serverConn.Write(message)
			if err != nil {
				log.Println("Error writing to server:", err.Error())
			}

			// wait for a response from the server
			log.Println("TO_SERVER: Waiting for a response from the server application")
			buf := make([]byte, 1024)
			reqLen, err := serverConn.Read(buf)
			if err != nil {
				log.Println("Error reading from server:", err.Error())
			}

			response := buf[:reqLen]

			// send the response to the relay server
			log.Println("TO_SERVER: Sending response to the relay server")
			responseChan <- response
		}
	}(responseChan, messageChan)

	<-done
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <relayAddress> <serverAddress>\n", os.Args[0])
	flag.PrintDefaults()
}
