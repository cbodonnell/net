package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/pion/dtls/v2"
)

var realAddr = &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5555}

var clients = make(map[string]net.Conn)
var lock sync.RWMutex

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
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Listening on %s\n", addr)

	// Dial to the real server
	realConn, err := net.DialUDP("udp", nil, realAddr)
	if err != nil {
		panic(err)
	}
	defer realConn.Close()

	fmt.Printf("Dialing to %s\n", realAddr)

	go func() {
		for {
			clientConn, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			fmt.Printf("Accepted connection from %s\n", clientConn.RemoteAddr())

			// forward the connection to the real server
			go registerConnection(clientConn, realConn)
		}
	}()

	go func() {
		for {
			if err := broadcastReal(realConn); err != nil {
				fmt.Printf("Error broadcasting real: %s\n", err)
				return
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	fmt.Println("Shutting down...")
}

func registerConnection(clientConn net.Conn, realConn *net.UDPConn) {
	lock.Lock()
	fmt.Printf("Registering connection from %s\n", clientConn.RemoteAddr())
	clients[clientConn.RemoteAddr().String()] = clientConn
	defer lock.Unlock()

	go func() {
		for {
			if err := forwardClient(clientConn, realConn); err != nil {
				fmt.Printf("Error forwarding client: %s\n", err)
				unregisterConnection(clientConn)
				return
			}
		}
	}()
}

func unregisterConnection(clientConn net.Conn) {
	lock.Lock()
	fmt.Printf("Unregistering connection from %s\n", clientConn.RemoteAddr())
	delete(clients, clientConn.RemoteAddr().String())
	lock.Unlock()
}

func forwardClient(clientConn net.Conn, realConn *net.UDPConn) error {
	buf := make([]byte, 1024)
	n, err := clientConn.Read(buf)
	if err != nil {
		return err
	}

	_, err = realConn.Write(buf[:n])
	if err != nil {
		return err
	}

	return nil
}

// broadcast real
func broadcastReal(realConn *net.UDPConn) error {
	buf := make([]byte, 1024)
	n, err := realConn.Read(buf)
	if err != nil {
		return err
	}

	lock.RLock()
	defer lock.RUnlock()
	for _, clientConn := range clients {
		_, err = clientConn.Write(buf[:n])
		if err != nil {
			fmt.Printf("Error broadcasting to client: %s\n", err)
			go unregisterConnection(clientConn)
		}
	}

	return nil
}
