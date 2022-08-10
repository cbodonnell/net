package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var tcpAddress string

func copyWorker(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	io.Copy(dst, src)
	doneCh <- true
}

func handleRequest(src net.Conn) {
	dst, err := net.Dial("tcp", tcpAddress)
	if err != nil {
		log.Println("dial:", err.Error())
		return
	}

	doneCh := make(chan bool)

	go copyWorker(dst, src, doneCh)
	go copyWorker(src, dst, doneCh)

	<-doneCh
	dst.Close()
	src.Close()
	<-doneCh
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <tcpTargetAddress>\n", os.Args[0])
	flag.PrintDefaults()
}

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

	// Listen on TCP port 2000 on all available unicast and
	// anycast IP addresses of the local system.
	l, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go handleRequest(conn)
	}
}
