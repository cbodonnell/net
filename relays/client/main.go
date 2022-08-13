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

var tcpAddress string
var key = []byte("passphrasewhichneedstobe32bytes!")

func main() {
	var port uint
	// var certFile string
	// var keyFile string

	flag.UintVar(&port, "p", 2222, "The port to listen on")
	flag.UintVar(&port, "port", 2222, "The port to listen on")
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

	l, err := net.Listen("tcp", portString)
	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}
	defer l.Close()

	log.Printf("Listening on %s\n", portString)
	log.Printf("Relaying to %s\n", tcpAddress)

	go func() {
		for {
			// Wait for a connection.
			clientConn, err := l.Accept()
			if err != nil {
				log.Fatal("Error accepting: ", err.Error())
			}

			// Handle the connection in a new goroutine.
			// The loop then returns to accepting, so that
			// multiple connections may be served concurrently.
			err = handleRequest(clientConn)
			if err != nil {
				log.Println("Error handling request:", err.Error())
				log.Println("Restarting listener in 1 second")
				time.Sleep(time.Second)
			}
		}
	}()

	// Wait for a signal to quit.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <tcpTargetAddress>\n", os.Args[0])
	flag.PrintDefaults()
}

func handleRequest(clientConn net.Conn) error {
	defer clientConn.Close()

	// read data from the client
	log.Println("Reading from client")
	buf := make([]byte, 1024)
	reqLen, err := clientConn.Read(buf)
	if err != nil {
		return err
	}

	msg := buf[:reqLen]

	log.Printf("Received message with %d bytes:\n%s\n", reqLen, string(msg))

	msg, err = crypto.Encrypt(key, msg)
	if err != nil {
		return err
	}

	// connect to the relay server
	log.Println("Connecting to relay server")
	relayConn, err := net.Dial("tcp", tcpAddress)
	if err != nil {
		return err
	}
	defer relayConn.Close()

	// write data to the destination
	log.Println("Writing to relay server")
	_, err = relayConn.Write(msg)
	if err != nil {
		return err
	}

	// read response from the destination
	log.Println("Reading from relay server")
	buf = make([]byte, 1024)
	reqLen, err = relayConn.Read(buf)
	if err != nil {
		return err
	}

	response := buf[:reqLen]

	response, err = crypto.Decrypt(key, response)
	if err != nil {
		return err
	}

	log.Printf("Received response with %d bytes:\n%s\n", reqLen, string(response))

	// copy response to the source
	_, err = clientConn.Write(response)
	if err != nil {
		return err
	}

	return nil
}
