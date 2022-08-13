package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

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

	aesCipher, err := crypto.NewAESCipher(crypto.AESCipherOpts{Key: key})
	if err != nil {
		log.Fatal("Error creating cipher:", err.Error())
	}

	for {
		// Wait for a connection.
		clientConn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err.Error())
		}

		// Handle the connection in a new goroutine.
		go func() {
			err = handleRequest(clientConn, aesCipher)
			if err != nil {
				log.Println("Error handling request:", err.Error())
			}
		}()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <tcpTargetAddress>\n", os.Args[0])
	flag.PrintDefaults()
}

func handleRequest(clientConn net.Conn, cipher crypto.Cipher) error {
	defer clientConn.Close()

	// connect to the relay server
	log.Println("Connecting to relay server")
	relayConn, err := net.Dial("tcp", tcpAddress)
	if err != nil {
		return err
	}
	defer relayConn.Close()

	return cipher.EncryptRoundTrip(relayConn, clientConn)
}
