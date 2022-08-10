package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var tcpAddress string

func main() {
	var port uint
	// var certFile string
	// var keyFile string

	flag.UintVar(&port, "p", 4223, "The port to listen on")
	flag.UintVar(&port, "port", 4223, "The port to listen on")
	// flag.StringVar(&certFile, "tlscert", "", "TLS cert file path")
	// flag.StringVar(&keyFile, "tlskey", "", "TLS key file path")
	// flag.BoolVar(&binaryMode, "b", false, "Use binary frames instead of text frames")
	// flag.BoolVar(&binaryMode, "binary", false, "Use binary frames instead of text frames")
	flag.Usage = usage
	flag.Parse()

	tcpAddress = flag.Arg(0)
	if tcpAddress == "" {
		log.Fatal("No address specified")
	}

	portString := fmt.Sprintf(":%d", port)

	log.Printf("Listening on %s\n", portString)
	log.Printf("Relaying to %s\n", tcpAddress)

	l, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer l.Close()
	for {
		// Wait for a connection.
		clientConn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go handleRequest(clientConn)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <tcpTargetAddress>\n", os.Args[0])
	flag.PrintDefaults()
}

func handleRequest(clientConn net.Conn) {
	defer clientConn.Close()

	// read data from the client
	buf := make([]byte, 1024)
	reqLen, err := clientConn.Read(buf)
	if err != nil {
		log.Println("Error reading from client:", err.Error())
	}

	msg := buf[:reqLen]

	log.Printf("Received message with %d bytes:\n%s\n", reqLen, string(msg))

	// connect to the relay server
	relayConn, err := net.Dial("tcp", tcpAddress)
	if err != nil {
		log.Println("dial:", err.Error())
		return
	}
	defer relayConn.Close()

	// write data to the destination
	_, err = relayConn.Write(msg)
	if err != nil {
		log.Println("Error writing to relay:", err.Error())
	}

	// read response from the destination
	buf = make([]byte, 1024)
	reqLen, err = relayConn.Read(buf)
	if err != nil {
		log.Println("Error reading from relay:", err.Error())
	}

	response := buf[:reqLen]

	log.Printf("Received response with %d bytes:\n%s\n", reqLen, string(response))

	// copy response to the source
	_, err = clientConn.Write(response)
	if err != nil {
		log.Println("Error writing to client:", err.Error())
	}
}
