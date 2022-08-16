package main

import (
	"bufio"
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
	}
	signalAddr := os.Args[1]

	if len(os.Args) < 3 {
		log.Fatalln("Please provide a server name to punch to")
	}
	serverName := os.Args[2]

	signal, err := net.ResolveUDPAddr("udp4", signalAddr)
	if err != nil {
		log.Fatalf("Failed to resolve signal address: %s", err)
	}

	listen, _ := net.ListenUDP("udp", nil)
	defer listen.Close()

	fmt.Printf("Listening on %s\n", listen.LocalAddr().String())

	target, err := punch(listen, signal, serverName)
	if err != nil {
		log.Fatalf("Failed to punch: %s", err)
	}

	fmt.Printf("Punched to %s\n", target.String())

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalln("Failed to read from stdin: ", err)
		}
		text = strings.TrimSpace(text)
		data := []byte(text)

		listen.SetDeadline(time.Now().Add(time.Second * 5))
		_, err = listen.WriteTo(data, target)
		if strings.TrimSpace(string(data)) == "STOP" {
			fmt.Println("Exiting UDP client!")
			return
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		buffer := make([]byte, 1024)
		n, remoteAddr, err := listen.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Reply from %s:\n%s\n", remoteAddr.String(), string(buffer[:n]))
	}
}

func punch(listen *net.UDPConn, signal *net.UDPAddr, name string) (*net.UDPAddr, error) {
	fmt.Printf("Punching to client %s on signal server %s\n", name, signal.String())

	listen.SetDeadline(time.Now().Add(time.Second * 5))
	defer listen.SetDeadline(time.Time{})

	_, err := listen.WriteTo([]byte(fmt.Sprintf("PUNCH: %s", name)), signal)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)
	n, _, err := listen.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}
	response := buffer[0:n]

	parts := strings.Split(string(response), ": ")
	if len(parts) != 2 {
		return nil, err
	}

	result, value := parts[0], parts[1]

	switch result {
	case "SUCCESS":
		target, _ := net.ResolveUDPAddr("udp4", value)
		return target, nil
	case "FAIL":
		return nil, fmt.Errorf("failed to punch to %s: %s", name, value)
	default:
		return nil, fmt.Errorf("unknown result: %s", result)
	}
}
