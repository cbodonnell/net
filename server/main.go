package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var port uint
	// var certFile string
	// var keyFile string

	flag.UintVar(&port, "p", 3333, "The port to listen on")
	flag.UintVar(&port, "port", 3333, "The port to listen on")
	// flag.StringVar(&certFile, "tlscert", "", "TLS cert file path")
	// flag.StringVar(&keyFile, "tlskey", "", "TLS key file path")
	// flag.BoolVar(&binaryMode, "b", false, "Use binary frames instead of text frames")
	// flag.BoolVar(&binaryMode, "binary", false, "Use binary frames instead of text frames")
	flag.Usage = usage
	flag.Parse()

	portString := fmt.Sprintf(":%d", port)

	log.Printf("Listening on %s\n", portString)

	// Listen on TCP port 2000 on all available unicast and
	// anycast IP addresses of the local system.
	l, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer l.Close()
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Close the connection when you're done with it.
	defer conn.Close()
	for {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 1024)
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err != nil {
			log.Println("Error reading:", err.Error())
			break
		}
		msg := buf[:reqLen]
		log.Printf("Received %d bytes from %s:\n%s\n", reqLen, conn.RemoteAddr().String(), string(msg))
		time.Sleep(time.Millisecond * 500)
		// Send a response back to person contacting us.
		_, err = conn.Write(msg)
		if err != nil {
			log.Println("Error writing:", err.Error())
			break
		}
	}
}
