package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/cbodonnell/tcp-queue/pkg/crypto"
)

type Server struct {
	relayAddress  string
	serverAddress string
	cipher        crypto.Cipher
	retryDuration time.Duration
	debug         bool
}

type ServerOpts struct {
	RelayAddress  string
	ServerAddress string
	Key           []byte
	RetryDuration string
	Debug         bool
}

func NewServer(opts ServerOpts) (*Server, error) {
	cipher, err := crypto.NewAESCipher(crypto.AESCipherOpts{Key: opts.Key})
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %s", err.Error())
	}

	retryDuration := time.Second
	if opts.RetryDuration != "" {
		retryDuration, err = time.ParseDuration(opts.RetryDuration)
		if err != nil {
			return nil, fmt.Errorf("error parsing retry duration: %s", err.Error())
		}
	}

	return &Server{
		relayAddress:  opts.RelayAddress,
		serverAddress: opts.ServerAddress,
		cipher:        cipher,
		retryDuration: retryDuration,
		debug:         opts.Debug,
	}, nil
}

func (s *Server) Run() error {
	go func() {
		for {
			if err := s.fetchAndRelay(); err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching and relaying: %s\n", err.Error())
				fmt.Fprintf(os.Stderr, "Retrying in %s\n", s.retryDuration)
				time.Sleep(s.retryDuration)
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	return errors.New("interrupted")
}

func (s *Server) fetchAndRelay() error {
	serverConn, err := net.Dial("tcp", s.serverAddress)
	if err != nil {
		return fmt.Errorf("error connecting to server: %s", err.Error())
	}
	if s.debug {
		log.Printf("Connected to server at %s\n", s.serverAddress)
	}
	defer serverConn.Close()

	relayConn, err := net.Dial("tcp", s.relayAddress)
	if err != nil {
		return fmt.Errorf("error connecting to relay: %s", err.Error())
	}
	if s.debug {
		log.Printf("Connected to relay at %s\n", s.relayAddress)
	}
	defer relayConn.Close()

	if s.debug {
		log.Println("Ready to relay")
	}

	return s.cipher.DecryptRoundTrip(serverConn, relayConn)
}
