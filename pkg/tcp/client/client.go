package client

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/cbodonnell/net/pkg/crypto"
)

type Client struct {
	port         uint
	relayAddress string
	cipher       crypto.Cipher
	bufferSize   uint
	debug        bool
}

type TCPClientOpts struct {
	Port         uint
	RelayAddress string
	Key          []byte
	BufferSize   uint
	Debug        bool
}

func NewTCPClient(opts TCPClientOpts) (*Client, error) {
	cipher, err := crypto.NewAESCipher(crypto.AESCipherOpts{Key: opts.Key})
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %s", err.Error())
	}

	return &Client{
		port:         opts.Port,
		relayAddress: opts.RelayAddress,
		cipher:       cipher,
		bufferSize:   opts.BufferSize,
		debug:        opts.Debug,
	}, nil
}

func (c *Client) Run() error {
	portString := fmt.Sprintf(":%d", c.port)
	if c.debug {
		log.Printf("Listening for client on %s\n", portString)
	}

	listener, err := net.Listen("tcp", portString)
	if err != nil {
		return err
	}
	defer listener.Close()

	errChan := make(chan error)

	go c.handleClientConnections(listener, errChan)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
		return errors.New("interrupted")
	case err := <-errChan:
		return fmt.Errorf("error handling client connections: %s", err.Error())
	}
}

func (c *Client) handleClientConnections(listener net.Listener, errChan chan<- error) error {
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			errChan <- fmt.Errorf("error accepting from client: %s", err.Error())
		}
		if err := c.handleRequest(clientConn); err != nil {
			fmt.Fprintf(os.Stderr, "error handling request: %s\n", err.Error())
		}
	}
}

func (c *Client) handleRequest(clientConn net.Conn) error {
	defer clientConn.Close()

	if c.debug {
		log.Printf("Connecting to %s\n", c.relayAddress)
	}
	relayConn, err := net.Dial("tcp", c.relayAddress)
	if err != nil {
		return err
	}
	defer relayConn.Close()

	if c.debug {
		log.Printf("Relaying to %s\n", c.relayAddress)
	}

	return c.cipher.EncryptRoundTrip(relayConn, clientConn)
}
