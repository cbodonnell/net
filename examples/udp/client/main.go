package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Please provide a host:port string")
		return
	}
	CONNECT := os.Args[1]

	s, err := net.ResolveUDPAddr("udp4", CONNECT)
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
	defer c.Close()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		data := []byte(text + "\n")
		_, err = c.Write(data)
		if strings.TrimSpace(string(data)) == "STOP" {
			fmt.Println("Exiting UDP client!")
			return
		}

		if err != nil {
			log.Fatal(err)
		}

		buffer := make([]byte, 1024)
		n, _, err := c.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Reply: %s\n", string(buffer[0:n]))
	}
}
