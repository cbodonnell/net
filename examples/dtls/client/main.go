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

var listenRealAddr = &net.UDPAddr{IP: net.ParseIP(""), Port: 3333}

var clients = make(map[string]*net.UDPAddr)
var lock sync.RWMutex

func main() {
	// Prepare the IP to connect to
	targetAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}

	// Read the client private key from disk
	clientKeyBytes, err := ioutil.ReadFile(path.Join("examples/x509/client", "key.pem"))
	if err != nil {
		panic(err)
	}

	// clientKeyPem, _ := pem.Decode(clientKeyBytes)

	// Read the client certificate from disk
	clientCertBytes, err := ioutil.ReadFile(path.Join("examples/x509/client", "cert.pem"))
	if err != nil {
		panic(err)
	}

	// clientCertPem, _ := pem.Decode(clientCertBytes)

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	targetConn, err := dtls.DialWithContext(ctx, "udp", targetAddr, config)
	if err != nil {
		panic(err)
	}
	defer targetConn.Close()

	fmt.Printf("Connected to %s\n", targetConn.RemoteAddr())

	// Listen for incoming connections from the real client
	realConn, err := net.ListenUDP("udp", listenRealAddr)
	if err != nil {
		panic(err)
	}
	defer realConn.Close()

	fmt.Printf("Listening on %s\n", realConn.LocalAddr())

	go func() {
		for {
			if err := forwardReal(realConn, targetConn); err != nil {
				fmt.Printf("Error round-tripping real: %s\n", err)
				return
			}
		}
	}()

	go func() {
		for {
			if err := broadcastTarget(targetConn, realConn); err != nil {
				fmt.Printf("Error round-tripping target: %s\n", err)
				return
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	fmt.Println("Shutting down...")
}

func forwardReal(realConn *net.UDPConn, targetConn *dtls.Conn) error {
	// read from the real client and send to the target
	buf := make([]byte, 1024)
	n, realAddr, err := realConn.ReadFromUDP(buf)
	if err != nil {
		return err
	}

	fmt.Printf("Received %d bytes from %s\n", n, realAddr.String())

	lock.Lock()
	defer lock.Unlock()
	if _, ok := clients[realAddr.String()]; !ok {
		fmt.Printf("Registering connection: %s\n", realAddr.String())
		clients[realAddr.String()] = realAddr
	}

	// send to target
	_, err = targetConn.Write(buf[:n])
	if err != nil {
		return err
	}

	return nil
}

func broadcastTarget(targetConn *dtls.Conn, realConn *net.UDPConn) error {
	// read from the target and send to the real client
	buf := make([]byte, 1024)
	n, err := targetConn.Read(buf)
	if err != nil {
		return err
	}

	fmt.Printf("Received %d bytes from %s\n", n, targetConn.RemoteAddr())

	lock.RLock()
	defer lock.RUnlock()
	for _, realAddr := range clients {
		fmt.Printf("Broadcasting %d bytes to %s\n", n, realAddr.String())
		_, err = realConn.WriteToUDP(buf[:n], realAddr)
		if err != nil {
			fmt.Printf("Error broadcasting to client: %s\n", err)
			go unregisterConnection(realAddr.String())
		}
	}

	return nil
}

func unregisterConnection(realAddr string) {
	lock.Lock()
	defer lock.Unlock()
	fmt.Printf("Unregistering connection: %s\n", realAddr)
	delete(clients, realAddr)
}
