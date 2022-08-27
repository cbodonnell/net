package server

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cbodonnell/net/pkg/crypto"
)

type UDPServer struct {
	relayAddress  string
	serverAddress string
	serverName    string
	cipher        crypto.Cipher
	retryDuration time.Duration
	debug         bool
}

type UDPServerOpts struct {
	RelayAddress  string
	ServerAddress string
	ServerName    string
	Key           []byte
	RetryDuration string
	Debug         bool
}

func NewUDPServer(opts UDPServerOpts) (*UDPServer, error) {
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

	return &UDPServer{
		relayAddress:  opts.RelayAddress,
		serverAddress: opts.ServerAddress,
		serverName:    opts.ServerName,
		cipher:        cipher,
		retryDuration: retryDuration,
		debug:         opts.Debug,
	}, nil
}

func (s *UDPServer) Run() error {
	relayAddr, err := net.ResolveUDPAddr("udp", s.relayAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve remote address: %s", err)
	}

	go func() {
		for {
			if err := s.registerAndServe(relayAddr); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to register and serve: %s", err.Error())
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

func (s *UDPServer) registerAndServe(relayAddr *net.UDPAddr) error {
	listen, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	defer listen.Close()

	fmt.Printf("Listening on %s\n", listen.LocalAddr().String())

	registerChan := make(chan string)
	pongChan := make(chan struct{})

	// Read messages from the relay server
	go func(registerChan chan<- string, pongChan chan<- struct{}) {
		for {
			buffer := make([]byte, 1024)
			n, remoteAddr, err := listen.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("[ERROR] Failed to read from UDP: %s\n", err.Error())
				continue
			}
			message := buffer[:n]

			fmt.Printf("[INCOMING] from %s:\n%s\n", remoteAddr.String(), string(message))

			switch remoteAddr.String() {
			case relayAddr.String():
				if err := handleRelayServerMessage(message, registerChan, pongChan); err != nil {
					fmt.Printf("failed to handle relay server message: %s\n", err)
					continue
				}
			default:
				// Handle other messages
				// TODO: Change this to a roundtrip to the server application (dial)
				_, err = listen.WriteToUDP(message, remoteAddr)
				if err != nil {
					fmt.Printf("[ERROR] Failed to write to %s: %s\n", remoteAddr.String(), err.Error())
					continue
				}
			}

		}
	}(registerChan, pongChan)

	// Register with the relay server
	relayErrChan := make(chan error)

	for {
		fmt.Printf("Registering with relay server %s\n", relayAddr.String())
		go func(relayErrChan chan<- error) {
			err := s.register(listen, relayAddr)
			if err != nil {
				relayErrChan <- fmt.Errorf("failed to register: %s", err.Error())
				return
			}

			select {
			case <-time.After(time.Second * 5):
				relayErrChan <- errors.New("registration timeout")
				return
			case target := <-registerChan:
				fmt.Printf("Registered as %s\n", target)
			}

			for {
				time.Sleep(time.Second * 5)
				if err := s.ping(listen, relayAddr); err != nil {
					relayErrChan <- fmt.Errorf("failed to ping: %s", err)
					return
				}
				select {
				case <-time.After(time.Second * 5):
					relayErrChan <- errors.New("ping timeout")
					return
				case <-pongChan:
				}
			}
		}(relayErrChan)
		err := <-relayErrChan
		fmt.Printf("failed to connect to relay server: %s\n", err.Error())
		time.Sleep(time.Second * 5)
	}
}

func (s *UDPServer) register(listen *net.UDPConn, remoteAddr *net.UDPAddr) error {
	_, err := listen.WriteTo([]byte(fmt.Sprintf("REGISTER: %s", s.serverName)), remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write to relay server %s: %s", remoteAddr.String(), err.Error())
	}

	return nil
}

func (s *UDPServer) ping(listen *net.UDPConn, remoteAddr *net.UDPAddr) error {
	_, err := listen.WriteTo([]byte(fmt.Sprintf("PING: %s", s.serverName)), remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write to relay server %s: %s", remoteAddr.String(), err.Error())
	}

	return nil
}

func handleRelayServerMessage(message []byte, registerChan chan<- string, pongChan chan<- struct{}) error {
	parts := strings.Split(string(message), ": ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid message from relay server: %s", string(message))
	}
	action, target := parts[0], parts[1]

	switch action {
	case "SUCCESS":
		registerChan <- target
	case "PONG":
		pongChan <- struct{}{}
	case "FAIL":
		return fmt.Errorf("failure message from relay server: %s", target)
	default:
		return fmt.Errorf("unknown action from relay server: %s", action)
	}

	return nil
}
