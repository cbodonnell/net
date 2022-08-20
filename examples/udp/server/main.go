package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Please provide a signal server address")
		return
	}
	signalAddr := os.Args[1]

	if len(os.Args) < 3 {
		log.Fatalln("Please provide a name")
		return
	}
	name := os.Args[2]

	signal, err := net.ResolveUDPAddr("udp", signalAddr)
	if err != nil {
		log.Fatalf("Failed to resolve remote address: %s", err)
	}

	if err := registerAndServe(signal, name); err != nil {
		log.Fatalf("Failed to register and serve: %s", err)
	}
}

func registerAndServe(signalAddr *net.UDPAddr, name string) error {
	listen, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	defer listen.Close()

	fmt.Printf("Listening on %s\n", listen.LocalAddr().String())

	registerChan := make(chan string)
	pongChan := make(chan struct{})

	// Read messages
	go func(registerChan chan<- string, pongChan chan<- struct{}) {
		for {
			buffer := make([]byte, 1024)
			bytesRead, remoteAddr, err := listen.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("[ERROR] Failed to read from UDP: %s\n", err.Error())
				continue
			}
			message := buffer[:bytesRead]

			fmt.Printf("[INCOMING] from %s:\n%s\n", remoteAddr.String(), string(message))

			switch remoteAddr.String() {
			case signalAddr.String():
				if err := handleSignalServerMessage(message, registerChan, pongChan); err != nil {
					fmt.Printf("failed to handle signal server message: %s\n", err)
					continue
				}
			default:
				// Handle other messages
				_, err = listen.WriteToUDP(message, remoteAddr)
				if err != nil {
					fmt.Printf("[ERROR] Failed to write to %s: %s\n", remoteAddr.String(), err.Error())
					continue
				}
			}

		}
	}(registerChan, pongChan)

	// Register with the signal server
	signalErrChan := make(chan error)

	for {
		fmt.Printf("Registering with signal server %s\n", signalAddr.String())
		go func(signalErrChan chan<- error) {
			err := register(listen, signalAddr, name)
			if err != nil {
				signalErrChan <- fmt.Errorf("failed to register: %s", err.Error())
				return
			}

			select {
			case <-time.After(time.Second * 5):
				signalErrChan <- errors.New("registration timeout")
				return
			case target := <-registerChan:
				fmt.Printf("Registered as %s\n", target)
			}

			for {
				time.Sleep(time.Second * 5)
				if err := ping(listen, signalAddr, name); err != nil {
					signalErrChan <- fmt.Errorf("failed to ping: %s", err)
					return
				}
				select {
				case <-time.After(time.Second * 5):
					signalErrChan <- errors.New("ping timeout")
					return
				case <-pongChan:
				}
			}
		}(signalErrChan)
		err := <-signalErrChan
		fmt.Printf("failed to connect to signal server: %s\n", err.Error())
		time.Sleep(time.Second * 5)
	}
}

func register(listen *net.UDPConn, remoteAddr *net.UDPAddr, name string) error {
	_, err := listen.WriteTo([]byte(fmt.Sprintf("REGISTER: %s", name)), remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write to signal server %s: %s", remoteAddr.String(), err.Error())
	}

	return nil
}

func ping(listen *net.UDPConn, remoteAddr *net.UDPAddr, name string) error {
	_, err := listen.WriteTo([]byte(fmt.Sprintf("PING: %s", name)), remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write to signal server %s: %s", remoteAddr.String(), err.Error())
	}

	return nil
}

func handleSignalServerMessage(message []byte, registerChan chan<- string, pongChan chan<- struct{}) error {
	parts := strings.Split(string(message), ": ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid message from signal server: %s", string(message))
	}
	action, target := parts[0], parts[1]

	switch action {
	case "SUCCESS":
		registerChan <- target
	case "PONG":
		pongChan <- struct{}{}
	case "FAIL":
		return fmt.Errorf("failure message from signal server: %s", target)
	default:
		return fmt.Errorf("unknown action from signal server: %s", action)
	}

	return nil
}
