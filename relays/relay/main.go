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

	done := make(chan bool)

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

		// responseQueue := make(chan chan []byte)

		for {
			// Listen for an incoming connections from the server.
			conn, err := serverListener.Accept()
			if err != nil {
				log.Fatal("Error accepting from server: ", err.Error())
			}
			// Handle connections from the server in a new goroutine.
			go handleServerRequest(conn, messageQueue)
		}
	}(serverListener, messageQueue)

	<-done
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
	log.Println("CLIENT: Reading from client")
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
	log.Printf("CLIENT: Received message with %d bytes:\n%s\n", reqLen, string(message.Data))

	// add message to the queue
	log.Println("CLIENT: Adding message to queue")
	queue <- message

	// wait for a response to the message
	log.Println("CLIENT: Waiting for a response")
	response := <-message.Response

	// Send a response back to person contacting us.
	log.Println("CLIENT: Sending response to client")
	_, err = conn.Write(response)
	if err != nil {
		log.Println("Error writing:", err.Error())
	}
}

// Handles incoming requests from the server component.
func handleServerRequest(conn net.Conn, messageQueue <-chan Message) {
	// Close the connection when you're done with it.
	defer conn.Close()

	log.Println("SERVER: Reading from message queue")
	message := <-messageQueue

	log.Printf("SERVER: Found message with %d bytes:\n%s\n", len(message.Data), string(message.Data))

	// Send the message data to the server component.
	log.Println("SERVER: Sending message to server")
	_, err := conn.Write(message.Data)
	if err != nil {
		log.Println("Error writing to server:", err.Error())
	}

	// // add the response channel for this message to the reponse queue
	// log.Println("SERVER: Adding response channel to response queue")
	// responseQueue <- message.Response

	// Make a buffer to hold incoming data.
	log.Println("SERVER: Reading from server")
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading from client:", err.Error())
	}

	response := buf[:reqLen]
	log.Printf("SERVER: Received response with %d bytes:\n%s\n", reqLen, string(response))

	if bytes.Equal(response, []byte("init")) {
		// if the response is the init message, no need to send it back to the client
		log.Println("SERVER: Received init message, not sending response to client")
		return
	}
	// data was received from the server, respond to client
	// responseChan := <-responseQueue
	message.Response <- response
}
