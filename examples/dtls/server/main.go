package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net"
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

	// Read the client certificate from disk
	serverCertBytes, err := ioutil.ReadFile(path.Join("examples/x509/server", "cert.pem"))
	if err != nil {
		panic(err)
	}

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
			return context.WithDeadline(ctx, time.Now().Add(time.Second*30))
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

	log.Printf("Listening on %s\n", addr)

	// Dial to the real server
	realConn, err := net.DialUDP("udp", nil, realAddr)
	if err != nil {
		panic(err)
	}
	defer realConn.Close()

	log.Printf("Dialing to %s\n", realAddr)

	go func(realConn net.Conn) error {
		for {
			log.Printf("Reading from %s\n", realConn.RemoteAddr())
			buf := make([]byte, 1024)
			n, err := realConn.Read(buf)
			if err != nil {
				return err
			}

			broadcast(buf[:n])
		}
	}(realConn)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		log.Printf("Accepted connection from %s\n", clientConn.RemoteAddr())

		register(clientConn)

		go func() error {
			for {
				log.Printf("Reading from %s\n", clientConn.RemoteAddr())
				buf := make([]byte, 1024)
				_ = clientConn.SetReadDeadline(time.Now().Add(time.Second * 30))
				n, err := clientConn.Read(buf)
				if err != nil {
					unregister(clientConn)
					return err
				}

				log.Printf("Writing to %s\n", realConn.RemoteAddr())
				_, err = realConn.Write(buf[:n])
				if err != nil {
					return err
				}
			}
		}()
	}
}

func register(conn net.Conn) {
	lock.Lock()
	defer lock.Unlock()
	clients[conn.RemoteAddr().String()] = conn
	log.Println("Registered", conn.RemoteAddr())
}

func unregister(conn net.Conn) {
	lock.Lock()
	defer lock.Unlock()
	delete(clients, conn.RemoteAddr().String())
	err := conn.Close()
	if err != nil {
		log.Println("Failed to disconnect", conn.RemoteAddr(), err)
	} else {
		log.Println("Disconnected ", conn.RemoteAddr())
	}
}

func broadcast(msg []byte) {
	lock.RLock()
	defer lock.RUnlock()
	for _, conn := range clients {
		if _, err := conn.Write(msg); err != nil {
			log.Printf("Failed to write to %s\n", conn.RemoteAddr().String())
		}
	}
}
