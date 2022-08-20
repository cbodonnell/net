package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cbodonnell/tcp-queue/pkg/crypto"
)

type UDPClient struct {
	port         uint
	relayAddress string
	serverName   string
	cipher       crypto.Cipher
	bufferSize   uint
	debug        bool
}

type UDPClientOpts struct {
	Port         uint
	RelayAddress string
	ServerName   string
	Key          []byte
	BufferSize   uint
	Debug        bool
}

func NewUDPClient(opts UDPClientOpts) (*UDPClient, error) {
	cipher, err := crypto.NewAESCipher(crypto.AESCipherOpts{Key: opts.Key})
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %s", err.Error())
	}

	return &UDPClient{
		port:         opts.Port,
		relayAddress: opts.RelayAddress,
		serverName:   opts.ServerName,
		cipher:       cipher,
		bufferSize:   opts.BufferSize,
		debug:        opts.Debug,
	}, nil
}

func (c *UDPClient) Run() error {
	relayAddr, err := net.ResolveUDPAddr("udp4", c.relayAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve relay address: %s", err.Error())
	}

	relayListener, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("failed to listen for relay: %s", err.Error())
	}
	defer relayListener.Close()

	if c.debug {
		fmt.Printf("Listening for relay responses on %s\n", relayAddr.String())
	}

	portString := fmt.Sprintf(":%d", c.port)

	listenAddr, err := net.ResolveUDPAddr("udp4", portString)
	if err != nil {
		return fmt.Errorf("failed to resolve listen address: %s", err.Error())
	}

	clientListener, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen for client: %s", err.Error())
	}
	defer clientListener.Close()

	if c.debug {
		fmt.Printf("Listening for client requests on %s\n", listenAddr.String())
	}

	if c.debug {
		fmt.Printf("Punching to client %s on signal server %s\n", c.serverName, relayAddr.String())
	}

	target, err := c.punch(relayListener, relayAddr)
	if err != nil {
		return fmt.Errorf("failed to punch: %s", err.Error())
	}
	// TODO: Implement a fallback mechanism when punching fails

	if c.debug {
		fmt.Printf("Punched to target %s\n", target.String())
	}

	go c.handleClientConnections(clientListener, relayListener, target)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	return errors.New("interrupted")
}

func (c *UDPClient) punch(relayListener *net.UDPConn, relayAddr *net.UDPAddr) (*net.UDPAddr, error) {
	_, err := relayListener.WriteTo([]byte(fmt.Sprintf("PUNCH: %s", c.serverName)), relayAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to write to relay: %s", err.Error())
	}

	relayListener.SetReadDeadline(time.Now().Add(time.Second * 5))

	buffer := make([]byte, 1024)
	n, _, err := relayListener.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read from relay: %s", err.Error())
	}
	response := buffer[0:n]

	parts := strings.Split(string(response), ": ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid response from relay: %s", response)
	}

	result, value := parts[0], parts[1]

	switch result {
	case "SUCCESS":
		target, err := net.ResolveUDPAddr("udp4", value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve target address: %s", err.Error())
		}
		return target, nil
	case "FAIL":
		return nil, fmt.Errorf("failed to punch to %s: %s", c.serverName, value)
	default:
		return nil, fmt.Errorf("unknown result: %s", result)
	}
}

func (c *UDPClient) handleClientConnections(clientListener, targetListener *net.UDPConn, target *net.UDPAddr) {
	for {
		if err := c.handleRequest(clientListener, targetListener, target); err != nil {
			fmt.Fprintf(os.Stderr, "error handling request: %s\n", err.Error())
		}
	}
}

func (c *UDPClient) handleRequest(clientListener, targetListener *net.UDPConn, target *net.UDPAddr) error {
	buffer := make([]byte, c.bufferSize)
	n, remoteAddr, err := clientListener.ReadFromUDP(buffer)
	if err != nil {
		return fmt.Errorf("failed to read from client: %s", err.Error())
	}

	message := buffer[:n]

	if c.debug {
		fmt.Printf("Received %d bytes from %s:\n%s\n", n, remoteAddr.String(), string(message))
	}

	_, err = clientListener.WriteTo(message, target)
	if err != nil {
		return fmt.Errorf("failed to write to target: %s", err.Error())
	}

	buffer = make([]byte, c.bufferSize)
	n, remoteAddr, err = targetListener.ReadFromUDP(buffer)
	if err != nil {
		return fmt.Errorf("failed to read from target: %s", err.Error())
	}

	response := buffer[:n]

	_, err = clientListener.WriteTo(response, remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write to client: %s", err.Error())
	}

	return nil
}
