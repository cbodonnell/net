package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/examples/util"
)

func main() {
	// Prepare the IP to connect to
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4444}

	// Create parent context to cleanup handshaking connections on exit.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Read the client private key from disk
	serverKeyBytes, err := ioutil.ReadFile(path.Join("examples/x509/server", "key.pem"))
	if err != nil {
		panic(err)
	}

	// serverKeyPem, _ := pem.Decode(serverKeyBytes)

	// Read the client certificate from disk
	serverCertBytes, err := ioutil.ReadFile(path.Join("examples/x509/server", "cert.pem"))
	if err != nil {
		panic(err)
	}

	// serverCertPem, _ := pem.Decode(serverCertBytes)

	serverCert, err := tls.X509KeyPair(serverCertBytes, serverKeyBytes)
	if err != nil {
		panic(err)
	}

	// Read the root CA from disk
	rootCABytes, err := ioutil.ReadFile(path.Join("examples/x509/root", "cert.pem"))
	if err != nil {
		panic(err)
	}

	rootPem, _ := pem.Decode(rootCABytes)

	// Parse the root CA
	root, err := x509.ParseCertificate(rootPem.Bytes)
	if err != nil {
		panic(err)
	}

	// Add the root CA to the pool
	roots := x509.NewCertPool()
	roots.AddCert(root)

	// Prepare the configuration of the DTLS connection
	config := &dtls.Config{
		Certificates:         []tls.Certificate{serverCert},
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		// Create timeout context for accepted connection.
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(ctx, 30*time.Second)
		},
		ClientAuth: dtls.RequireAndVerifyClientCert,
		ClientCAs:  roots,
	}

	// Connect to a DTLS server
	listener, err := dtls.Listen("udp", addr, config)
	util.Check(err)
	defer func() {
		util.Check(listener.Close())
	}()

	fmt.Println("Listening")

	// Simulate a chat session
	hub := util.NewHub()

	go func() {
		for {
			// Wait for a connection.
			conn, err := listener.Accept()
			util.Check(err)
			// defer conn.Close() // TODO: graceful shutdown

			// `conn` is of type `net.Conn` but may be casted to `dtls.Conn`
			// using `dtlsConn := conn.(*dtls.Conn)` in order to to expose
			// functions like `ConnectionState` etc.

			// Register the connection with the chat hub
			if err == nil {
				hub.Register(conn)
			}
		}
	}()

	// Start chatting
	hub.Chat()
}
