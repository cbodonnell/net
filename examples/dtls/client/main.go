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
	"github.com/pion/udp"
)

var listenRealAddr = &net.UDPAddr{IP: net.ParseIP(""), Port: 3333}

var clients = make(map[string]net.Conn)
var lock sync.RWMutex

func main() {
	// Prepare the IP to connect to
	targetAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}

	// Read the client private key from disk
	clientKeyBytes, err := ioutil.ReadFile(path.Join("examples/x509/client", "key.pem"))
	if err != nil {
		panic(err)
	}

	// Read the client certificate from disk
	clientCertBytes, err := ioutil.ReadFile(path.Join("examples/x509/client", "cert.pem"))
	if err != nil {
		panic(err)
	}

	clientCert, err := tls.X509KeyPair(clientCertBytes, clientKeyBytes)
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
		Certificates: []tls.Certificate{clientCert},
		// InsecureSkipVerify:   true,
		RootCAs:              roots,
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
	}

	// Connect to a DTLS server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	targetConn, err := dtls.DialWithContext(ctx, "udp", targetAddr, config)
	if err != nil {
		panic(err)
	}
	defer targetConn.Close()

	log.Printf("Connected to %s\n", targetConn.RemoteAddr())

	// Listen for incoming connections from the real client
	listenReal, err := udp.Listen("udp", listenRealAddr)
	if err != nil {
		panic(err)
	}
	defer listenReal.Close()

	log.Printf("Listening on %s\n", listenReal.Addr())

	go func(targetConn net.Conn) error {
		for {
			log.Printf("Reading from %s\n", targetConn.RemoteAddr())
			buf := make([]byte, 1024)
			n, err := targetConn.Read(buf)
			if err != nil {
				return err
			}

			broadcast(buf[:n])
		}
	}(targetConn)

	for {
		realConn, err := listenReal.Accept()
		if err != nil {
			panic(err)
		}

		log.Printf("Accepted connection from %s\n", realConn.RemoteAddr())

		register(realConn)

		go func() error {
			for {
				log.Printf("Reading from %s\n", realConn.RemoteAddr())
				buf := make([]byte, 1024)
				_ = realConn.SetReadDeadline(time.Now().Add(time.Second * 30))
				n, err := realConn.Read(buf)
				if err != nil {
					unregister(realConn)
					return err
				}

				log.Printf("Writing to %s\n", targetConn.RemoteAddr())
				_, err = targetConn.Write(buf[:n])
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
		log.Printf("Writing to %s\n", conn.RemoteAddr())
		if _, err := conn.Write(msg); err != nil {
			log.Printf("Failed to write to %s\n", conn.RemoteAddr().String())
		}
	}
}
