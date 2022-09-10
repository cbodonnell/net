package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a host:port string")
		return
	}
	CONNECT := arguments[1]

	s, err := net.ResolveUDPAddr("udp4", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	fmt.Printf("Connecting to %s\n", s.String())

	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := c.Read(buffer)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("-> %s\n", string(buffer[0:n]))
		}
	}()

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			data := []byte(strings.TrimSuffix(text, "\n"))
			_, err = c.Write(data)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	fmt.Println("Shutting down...")
}
