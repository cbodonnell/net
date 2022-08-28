package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}
	PORT := ":" + arguments[1]

	s, err := net.ResolveUDPAddr("udp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("The UDP echo server listening on %s\n", s.String())

	defer connection.Close()
	buffer := make([]byte, 1024)
	rand.Seed(time.Now().Unix())

	for {
		n, addr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("-> %s\n", string(buffer[0:n]))

		_, err = connection.WriteToUDP(buffer[0:n], addr)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
