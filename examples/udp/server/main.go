package main

import (
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

	remote, err := net.ResolveUDPAddr("udp", signalAddr)
	if err != nil {
		log.Fatalf("Failed to resolve remote address: %s", err)
	}
	// local, err := net.ResolveUDPAddr("udp", localAddr)
	// if err != nil {
	// 	log.Fatalf("Failed to resolve local address: %s", err)
	// }
	listen, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}
	defer listen.Close()

	if err = register(listen, remote, name); err != nil {
		log.Fatalf("Failed to register: %s", err)
	}

	go func() {
		for {
			if err = ping(listen, remote, name); err != nil {
				log.Fatalf("Failed to ping: %s", err)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	fmt.Printf("Listening on %s\n", listen.LocalAddr().String())

	for {
		buffer := make([]byte, 1024)
		bytesRead, remoteAddr, err := listen.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("[ERROR] Failed to read from UDP: %s\n", err.Error())
			continue
		}
		message := buffer[:bytesRead]

		fmt.Printf("[INCOMING] from %s:\n%s\n", remoteAddr.String(), string(message))

		_, err = listen.WriteToUDP(message, remoteAddr)
		if err != nil {
			fmt.Printf("[ERROR] Failed to write to %s: %s\n", remoteAddr.String(), err.Error())
			continue
		}
	}
}

func register(listen *net.UDPConn, remoteAddr *net.UDPAddr, name string) error {
	fmt.Printf("Registering with signal server %s\n", remoteAddr.String())

	listen.SetDeadline(time.Now().Add(time.Second * 5))
	defer listen.SetDeadline(time.Time{})

	_, err := listen.WriteTo([]byte(fmt.Sprintf("REGISTER: %s", name)), remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write to signal server %s: %s", remoteAddr.String(), err.Error())
	}

	buffer := make([]byte, 1024)
	n, _, err := listen.ReadFromUDP(buffer)
	if err != nil {
		return fmt.Errorf("failed to read from signal server %s: %s", remoteAddr.String(), err.Error())
	}

	response := string(buffer[:n])

	parts := strings.Split(response, ": ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid response from signal server %s: %s", remoteAddr.String(), response)
	}
	action, target := parts[0], parts[1]

	switch action {
	case "SUCCESS":
		fmt.Printf("Registered as %s\n", target)
		return nil
	case "FAIL":
		return fmt.Errorf("failed to register as %s: %s", name, target)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func ping(listen *net.UDPConn, remoteAddr *net.UDPAddr, name string) error {
	listen.SetDeadline(time.Now().Add(time.Second * 5))
	defer listen.SetDeadline(time.Time{})

	_, err := listen.WriteTo([]byte(fmt.Sprintf("PING: %s", name)), remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to write to signal server %s: %s", remoteAddr.String(), err.Error())
	}
	// TODO: Need to read something from here to make sure the ping was successful
	// Maybe can add a switch statement on thew remoteAddr.String() to handle different senders

	return nil
}
