package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

type Message struct {
	Data     []byte
	Response chan []byte
}

func main() {
	var clientPort uint
	var serverPort uint
	// var certFile string
	// var keyFile string

	flag.UintVar(&clientPort, "client-port", 3333, "The port to listen for the client on")
	flag.UintVar(&serverPort, "server-port", 4444, "The port to listen for the server on")
	// flag.StringVar(&certFile, "tlscert", "", "TLS cert file path")
	// flag.StringVar(&keyFile, "tlskey", "", "TLS key file path")
	// flag.BoolVar(&binaryMode, "b", false, "Use binary frames instead of text frames")
	// flag.BoolVar(&binaryMode, "binary", false, "Use binary frames instead of text frames")
	flag.Usage = usage
	flag.Parse()

	clientPortString := fmt.Sprintf(":%d", clientPort)
	log.Printf("Listening for client on %s\n", clientPortString)

	clientListener, err := net.Listen("tcp", clientPortString)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}

	messageQueue := make(chan Message)

	go func(clientListener net.Listener, messageQueue chan Message) {
		defer clientListener.Close()
		for {
			// Listen for an incoming connections from the client.
			conn, err := clientListener.Accept()
			if err != nil {
				log.Fatal("Error accepting from client: ", err.Error())
			}
			// Handle connections from the client in a new goroutine.
			go handleClientRequest(conn, messageQueue)
		}
	}(clientListener, messageQueue)

	serverPortString := fmt.Sprintf(":%d", serverPort)
	log.Printf("Listening for server on %s\n", serverPortString)

	serverListener, err := net.Listen("tcp", serverPortString)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}

	go func(serverListener net.Listener, messageQueue chan Message) {
		defer serverListener.Close()

		responseQueue := make(chan chan []byte)

		for {
			// Listen for an incoming connections from the server.
			conn, err := serverListener.Accept()
			if err != nil {
				log.Fatal("Error accepting from server: ", err.Error())
			}
			// Handle connections from the server in a new goroutine.
			go handleServerRequest(conn, messageQueue, responseQueue)
		}
	}(serverListener, messageQueue)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
	flag.PrintDefaults()
}

// Handles incoming requests from the client component.
func handleClientRequest(conn net.Conn, queue chan<- Message) {
	// Close the connection when you're done with it.
	defer conn.Close()

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading from client:", err.Error())
	}

	message := Message{
		Data:     buf[:reqLen],
		Response: make(chan []byte),
	}
	log.Printf("Received message with %d bytes:\n%s\n", reqLen, string(message.Data))

	// add message to the queue
	queue <- message

	// wait for a response to the message
	response := <-message.Response

	// Send a response back to person contacting us.
	_, err = conn.Write(response)
	if err != nil {
		log.Println("Error writing:", err.Error())
	}
}

// Handles incoming requests from the server component.
func handleServerRequest(conn net.Conn, messageQueue <-chan Message, responseQueue chan chan []byte) {
	// Close the connection when you're done with it.
	defer conn.Close()

	message := <-messageQueue

	log.Printf("Found message with %d bytes:\n%s\n", len(message.Data), string(message.Data))

	// Send the message data to the server component.
	_, err := conn.Write(message.Data)
	if err != nil {
		log.Println("Error writing to server:", err.Error())
	}

	// add the response channel for this message to the reponse queue
	responseQueue <- message.Response

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading from client:", err.Error())
	}

	response := buf[:reqLen]
	log.Printf("Received response with %d bytes:\n%s\n", reqLen, string(response))

	if bytes.Equal(response, []byte("init")) {
		// if the response is the init message, no need to send it back to the client
		return
	}
	// data was received from the server, respond to client
	responseChan := <-responseQueue
	responseChan <- response
}
